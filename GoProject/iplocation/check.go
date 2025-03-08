package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"

	// "time"
	"util"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type Dblocation struct {
	code, ip, port     string
	changeip_period    int
	last_changeip_time string
	location           string
}

var Db *sql.DB

var (
	logFile  = "/data/kdl/log/devops/dps_location_check.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var dbPath = "/data/ip2region/maker/golang/ip2region.xdb"

// var newTime = time.Now().Add(-time.Duration(15) * time.Second)

var searcher *xdb.Searcher

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: caiweitao")
		fmt.Println("Version: 1.0.1")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2023-09-11")
		return
	}

	mess := []string{}
	data := make(chan string, 5000)
	node := getDb()
	done := make(chan struct{})
	numWorkers := len(node)

	go func() {
		for i := 0; i < numWorkers; i++ {
			<-done
		}
		close(data)
	}()

	for _, no := range node {
		go func(no Dblocation) {
			checkDpsLocation(no, searcher, data)
			done <- struct{}{}
		}(no)
	}

	for info := range data {
		mess = append(mess, info)
	}
	if len(mess) > 0 {
		util.SendMess(mess)
	}
	defer searcher.Close()
	logInfo.Printf("Totle To Be Detected Dps: %d", len(node))

}

func init() {
	var err error
	Db, err = util.ConnDb("root", "", "db")
	if err != nil {
		logError.Fatalf("ConnDb Failed : %s", err)
	}
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		logError.Fatalf("failed to load content from `%s`: %s\n", dbPath, err)
	}
	searcher, err = xdb.NewWithBuffer(cBuff)
	if err != nil {
		logError.Fatalf("failed to create searcher with content: %s\n", err)
	}
}

func getDb() []Dblocation {
	sqlStr := "select code,ip,port,changeip_period,last_changeip_time,location from dps where dps_type not in (2,6) and status = 1"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logError.Fatalf("Query Failed : %s", err)
	}
	info := make([]Dblocation, 0)
	defer rows.Close()
	for rows.Next() {
		var dps Dblocation
		err := rows.Scan(&dps.code, &dps.ip, &dps.port, &dps.changeip_period, &dps.last_changeip_time, &dps.location)
		if err != nil {
			logError.Fatalf("Scan Failed : %s", err)
		}
		// last_changeip_time, err := time.Parse("2006-01-02 15:04:05", dps.last_changeip_time)
		if err != nil {
			logError.Fatalf("time.Parse Failed : %s", err)
		}
		info = append(info, dps)
		// if last_changeip_time.Add(time.Duration(dps.changeip_period) * time.Second).Before(newTime) {
		// 	continue
		// } else {
		// 	info = append(info, dps)
		// }
	}
	return info
}

func checkDpsLocation(node Dblocation, searcher *xdb.Searcher, data chan string) {
	var location string
	region, err := searcher.SearchByStr(node.ip)
	if err != nil {
		fmt.Printf("failed to SearchIP(%s): %s\n", node.ip, err)
		return
	}
	st := strings.Split(region, "|")
	if st[2] == strings.Split(st[3], "市")[0] {
		location = st[3]
	} else {
		location = st[2] + st[3]
	}
	if location != node.location {
		data <- fmt.Sprintf("%s(%s) ----> %s", node.code, node.location, location)
	}
}
