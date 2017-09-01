package main

import (
	"flag"
	"fmt"
	"logBackup"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"strconv"
	"os/user"
)

var optionConfig = flag.String("config", "/etc/logBackup.conf", "config")
var optionIsServer = flag.Bool("server", false, "run as server mode")
var optionIsClient = flag.Bool("client", true, "run as client mode")
var optionClientFile = flag.String("file", "", "set send file to server")
var optionClientPath = flag.String("path", "", "set send file to server backup path")
var optionClientName = flag.String("name", "", "rename send file to server backup path")
var optionVerbose = flag.Bool("verbose", false, `show run details`)

func usage() {
	fmt.Printf("Usage: %s [options]\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	flag.Parse()

	if *optionVerbose {
		os.Setenv("DEBUG", "ok")
	}

	file, err := logBackup.LoadFile(*optionConfig)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if *optionIsServer {
		if _, err := os.Stat(*optionConfig); err != nil {
			fmt.Printf("Sorry,not found %s\n", *optionConfig)
			os.Exit(1)
		}

		addr, ok := file.Get("server", "bind-address")
		if !ok {
			fmt.Printf("Sorry,config file not right %s with bind-address\n", *optionConfig)
			os.Exit(1)
		}

		datadir, ok := file.Get("server", "datadir")
		if !ok {
			fmt.Printf("Sorry,config file not right %s with datadir\n", *optionConfig)
			os.Exit(1)
		}

		logBackup.GloablConfig.Addr = addr
		logBackup.GloablConfig.BackupPath = datadir

		func() {
			userStr, ok := file.Get("server", "user")
			if !ok {
				fmt.Printf("Sorry,config file not right %s with user\n", *optionConfig)
				return
			}

			u, err := user.Lookup(userStr)
			if err != nil {
				fmt.Printf("Sorry,not found user %s in your machine\n", userStr)
				return
			}

			uid, err := strconv.Atoi(u.Uid)
			if err != nil {
				fmt.Printf("Sorry,convert %s to type int failed %s \n", userStr, err)
				return
			}

			groupStr, ok := file.Get("server", "group")
			if !ok {
				fmt.Printf("Sorry,config file not right %s\n", *optionConfig)
				return
			}

			g, err := user.LookupGroup(groupStr)
			if err != nil {
				fmt.Printf("Sorry,not found group %s in your machine\n", groupStr)
				return
			}

			gid, err := strconv.Atoi(g.Gid)
			if err != nil {
				fmt.Printf("Sorry,convert %s to type int failed %s \n", userStr, err)
				return
			}

			logBackup.GloablConfig.ToChown = true
			logBackup.GloablConfig.Uid = uid
			logBackup.GloablConfig.Gid = gid
		}()

		logBackup.Debugf("server will chown by %d:%d", logBackup.GloablConfig.Uid, logBackup.GloablConfig.Uid)

		server, err := logBackup.NewServer()

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
			if matched, err := regexp.Match(`^[0-9a-zA-Z\-_]{1,}$`, []byte(saveRelativePath)); err != nil || !matched {
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

		if len(*optionClientName) > 0 {
			if matched, err := regexp.Match(`^[0-9a-zA-Z\-_.]{1,}$`, []byte(*optionClientName)); err != nil || !matched {
				fmt.Printf("Sorry, transfer file name is invalide\n")
				os.Exit(1)
			}
		}

		addr, ok := file.Get("client", "server-address")
		if !ok {
			fmt.Printf("Sorry,config file not right %s\n", *optionConfig)
			os.Exit(1)
		}

		logBackup.Transerf(addr, *optionClientFile, saveRelativePath, *optionClientName)
	}

	os.Exit(0)
}
