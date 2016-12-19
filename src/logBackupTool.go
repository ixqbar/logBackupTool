package main

import (
	"fmt"
	"os"
	"flag"
	"runtime"
	"logBackup"
	"os/signal"
	"syscall"
	"strings"
)

var optionConfig     = flag.String("c", "/etc/logBackup.conf", "config")
var optionIsServer   = flag.Bool("ms", false, "run as server mode")
var optionIsClient   = flag.Bool("mc", true, "run as client mode")
var optionClientFile = flag.String("file", "", "send file to server")
var optionClientPath = flag.String("path", "", "send file to server root path")
var optionVerbose    = flag.Bool("verbose", true, `show run details`)

func usage() {
	fmt.Printf("Usage: %s [options]\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main()  {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	flag.Parse()

	if *optionVerbose {
		os.Setenv("DEBUG", "ok")
	}

	file, err := logBackup.LoadFile(*optionConfig)
	if err != nil {
		fmt.Printf("Sorry,config file not right %s %v\n", *optionConfig, err)
		os.Exit(1)
	}

	if *optionIsServer {
		if _, err := os.Stat(*optionConfig); err != nil {
			fmt.Printf("Sorry,not found %s\n", *optionConfig)
			os.Exit(1)
		}

		addr, ok := file.Get("server", "bind-address")
		if !ok {
			fmt.Printf("Sorry,config file not right %s\n", *optionConfig)
			os.Exit(1)
		}

		datadir, ok := file.Get("server", "datadir")
		if !ok {
			fmt.Printf("Sorry,config file not right %s\n", *optionConfig)
			os.Exit(1)
		}

		if false == strings.HasSuffix(datadir, "/") {
			datadir = datadir + "/"
		}

		server, err := logBackup.NewServer(&logBackup.Config{
			Addr       : addr,
			BackupPath : datadir,
		})

		if err != nil {
			fmt.Printf("Sorry, start server failed %v\n", err)
			os.Exit(1)
		}

		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			<-sigs
			server.Stop()
		}()

		server.Start()
	}

	if *optionIsClient && false == *optionIsServer {
		if len(*optionClientFile) == 0 {
			fmt.Printf("Sorry, transfer file not empty\n")
			os.Exit(1)
		}

		saveRelativePath := *optionClientPath
		if len(saveRelativePath) > 0 {
			if strings.Index(saveRelativePath, ".") >= 0 {
				fmt.Printf("Sorry, transfer file path is invalide\n")
				os.Exit(1)
			}

			if strings.HasPrefix(saveRelativePath, "/") {
				saveRelativePath = saveRelativePath[1:]
			}

			if strings.HasSuffix(saveRelativePath, "/") == false {
				saveRelativePath = saveRelativePath + "/"
			}
		}

		addr, ok := file.Get("client", "server-address")
		if !ok {
			fmt.Printf("Sorry,config file not right %s\n", *optionConfig)
			os.Exit(1)
		}

		logBackup.Transerf(addr, *optionClientFile, saveRelativePath)
	}

	os.Exit(0)
}
