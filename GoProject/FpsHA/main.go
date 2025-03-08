package main

import (
	fps "FpsHA/fpsutil"
	"flag"
	"fmt"
	"time"
	"util"
)

var logFile = "/data/kdl/log/devops/FPS_HA.log"

var logInfo, logError = util.InitLog(logFile, "[INFO] ", "[ERROR] ")

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.5")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2023-11-30")
		return
	}
	start := time.Now()
	logInfo.Printf("Run FPS_HA ...")
	_err := fps.CheckNode()
	if _err != nil {
		logError.Printf("GetFps Failed : %s", _err)
	}
	end := time.Now()
	logInfo.Printf("Time Use: %v", end.Sub(start))
}
