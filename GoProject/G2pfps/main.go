package main

import (
	"flag"
	"fmt"
	fps "fps/fpslib"
	"time"
	"util"
)

var (
	logFile = "/data/kdl/log/G2pfpsmonitor.log"
	logInfo = util.LogConf(logFile, "[INFO] ")
)

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.11")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2024-08-09")
		return
	}
	var sleep time.Duration
	sleepTime := 30 * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		start := time.Now()
		node := fps.GetFpsNode()
		aliveList, deadList := fps.ProcessFpsNodes(node.Node)
		logInfo.Printf("alive fps node: %v, dead fps node: %v", aliveList, deadList)
		fps.Fpsreport(aliveList, deadList)

		end := time.Since(start)
		if int(end.Seconds()) <= 0 || int(end.Seconds()) >= 30 {
			sleep = 3 * time.Second
		} else {
			sleep = sleepTime - end
		}
		logInfo.Printf("Time Use: %v, sleep: %v\n", end, sleep)
		<-ticker.C
	}
}
