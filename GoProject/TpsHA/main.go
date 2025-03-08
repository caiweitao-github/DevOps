package main

import (
	tps "TpsHA/tpsutil"
	"flag"
	"fmt"
	"time"
	"util"
)

var (
	logFile  = "/data/kdl/log/devops/TPS_HA.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.7")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2023-11-29")
		return
	}
	start := time.Now()
	logInfo.Printf("Run TPS_HA ...")
	_err := tps.CheckNode()
	if _err != nil {
		logError.Printf("GetTps Failed : %s", _err)
	}
	tps.TerminateServer()
	end := time.Now()
	logInfo.Printf("Time Use: %v", end.Sub(start))
}
