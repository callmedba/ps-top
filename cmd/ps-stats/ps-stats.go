// ps-stats - vmstat like program which collects information from MySQL's
// performance_schema database.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sjmudd/ps-top/app"
	"github.com/sjmudd/ps-top/connector"
	"github.com/sjmudd/ps-top/lib"
	"github.com/sjmudd/ps-top/version"
)

var (
	count	int
	delay	int

	cpuprofile         = flag.String("cpuprofile", "", "write cpu profile to file")
	flag_debug         = flag.Bool("debug", false, "Enabling debug logging")
	flag_defaults_file = flag.String("defaults-file", "", "Provide a defaults-file to use to connect to MySQL")
	flag_help          = flag.Bool("help", false, "Provide some help for "+lib.MyName())
	flag_host          = flag.String("host", "", "Provide the hostname of the MySQL to connect to")
	flag_limit         = flag.Int("limit", 0, "Show a maximum of limit entries (defaults to screen size if output to screen)")
	flag_password      = flag.String("password", "", "Provide the password when connecting to the MySQL server")
	flag_port          = flag.Int("port", 0, "Provide the port number of the MySQL to connect to (default: 3306)") /* deliberately 0 here, defaults to 3306 elsewhere */
	flag_socket        = flag.String("socket", "", "Provide the path to the local MySQL server to connect to")
	flag_totals        = flag.Bool("totals", false, "Only show the totals when in stdout mode and no detail (default: false)")
	flag_user          = flag.String("user", "", "Provide the username to connect with to MySQL (default: $USER)")
	flag_version       = flag.Bool("version", false, "Show the version of "+lib.MyName())
	flag_view          = flag.String("view", "", "Provide view to show when starting pstop (default: table_io_latency)")
)

func usage() {
	fmt.Println(lib.MyName() + " - " + lib.Copyright())
	fmt.Println("")
	fmt.Println("vmstat-like program to show MySQL activity by using information collected")
	fmt.Println("from performance_schema without sent to stdout.")
	fmt.Println("")
	fmt.Println("Usage: " + lib.MyName() + " <options> [delay [count]]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("--defaults-file=/path/to/defaults.file   Connect to MySQL using given defaults-file")
	fmt.Println("--help                                   Show this help message")
	fmt.Println("--host=<hostname>                        MySQL host to connect to")
	fmt.Println("--limit=<rows>                           Limit the number of lines of output (excluding headers)")
	fmt.Println("--password=<password>                    Password to use when connecting")
	fmt.Println("--port=<port>                            MySQL port to connect to")
	fmt.Println("--socket=<path>                          MySQL path of the socket to connect to")
	fmt.Println("--totals                                 Only send the totals to stdout (in stdout mode)")
	fmt.Println("--user=<user>                            User to connect with")
	fmt.Println("--version                                Show the version")
	fmt.Println("--view=<view>                            Determine the view you want to see when pstop starts (default: table_io_latency")
	fmt.Println("                                         Possible values: table_io_latency table_io_ops file_io_latency table_lock_latency user_latency mutex_latency stages_latency")
}

func main() {
	var connector connector.Connector
	var defaults_file string = ""
	var err = errors.New("unknown")

	flag.Parse()

	// Too many arguments
	if len(flag.Args()) > 2 {
		usage()
		os.Exit(1)
	}
	// delay
	if len(flag.Args()) >= 1 {
		delay, err = strconv.Atoi(flag.Args()[0])
		if err != nil {
			log.Fatal("Unable to parse delay: ", err)
		}
	} else {
		delay = 1
	}
	// count
	if len(flag.Args()) >= 2 {
		count, err = strconv.Atoi(flag.Args()[1])
		if err != nil {
			log.Fatal("Unable to parse count: ", err)
		}
	} else {
		count = 0
	}

	// clean me up
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *flag_debug {
		lib.Logger.EnableLogging(true)
	}
	if *flag_version {
		fmt.Println(lib.MyName() + " version " + version.Version())
		return
	}
	if *flag_help {
		usage()
		return
	}

	lib.Logger.Println("Starting " + lib.MyName())

	if *flag_host != "" || *flag_socket != "" {
		lib.Logger.Println("--host= or --socket= defined")
		var components = make(map[string]string)
		if *flag_host != "" && *flag_socket != "" {
			fmt.Println(lib.MyName() + ": Do not specify --host and --socket together")
			os.Exit(1)
		}
		if *flag_host != "" {
			components["host"] = *flag_host
		}
		if *flag_port != 0 {
			if *flag_socket == "" {
				components["port"] = fmt.Sprintf("%d", *flag_port)
			} else {
				fmt.Println(lib.MyName() + ": Do not specify --socket and --port together")
				os.Exit(1)
			}
		}
		if *flag_socket != "" {
			components["socket"] = *flag_socket
		}
		if *flag_user != "" {
			components["user"] = *flag_user
		}
		if *flag_password != "" {
			components["password"] = *flag_password
		}
		connector.ConnectByComponents(components)
	} else {
		if flag_defaults_file != nil && *flag_defaults_file != "" {
			lib.Logger.Println("--defaults-file defined")
			defaults_file = *flag_defaults_file
		} else {
			lib.Logger.Println("connecting by implicit defaults file")
		}
		connector.ConnectByDefaultsFile(defaults_file)
	}

	var app app.App

	app.Setup(connector.Handle(), delay, count, true, *flag_limit, *flag_view, *flag_totals)
	app.Run()
	app.Cleanup()
	lib.Logger.Println("Terminating " + lib.MyName())
}
