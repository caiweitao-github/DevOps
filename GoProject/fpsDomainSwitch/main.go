package main

import (
	"Aliyun"
	"database/sql"
	"flag"
	"fmt"
	"regexp"
)

var DB *sql.DB

var domainName = "kdlfps.com"

type Data struct {
	id                 int
	code, location, ip string
	domainlist         []string
}

func main() {
	srcCode := flag.String("src_code", "", "源机器code")
	destCode := flag.String("dest_code", "", "目标机器code")
	flag.Parse()
	res, srcInfo, destInfo := checkLocation(*srcCode, *destCode)
	if res {
		switchDomain(srcInfo, destInfo)
	}
}

func init() {
	var err error
	DB, err = util.dbDB()
	if err != nil {
		fmt.Printf("ConnDb Failed : %s", err)
		return
	}
}

func checkLocation(srcCode, destCode string) (res bool, srcDtat, destData Data) {
	sqlStr := "select id,code,location_code,login_ip from fps where code = ? and status in (1, 2, 3)"
	sqlStr2 := "select id,code,location_code,login_ip from fps where code = ? and status in (1, 2, 3)"
	err := DB.QueryRow(sqlStr, srcCode).Scan(&srcDtat.id, &srcDtat.code, &srcDtat.location, &srcDtat.ip)
	if err != nil {
		fmt.Printf("query failed: %v\n", err)
		return
	}
	err1 := DB.QueryRow(sqlStr2, destCode).Scan(&destData.id, &destData.code, &destData.location, &destData.ip)
	if err1 != nil {
		fmt.Printf("query failed: %v\n", err)
		return
	}

	if srcDtat.location != destData.location {
		fmt.Printf("源目机器非同一地区: %s ---> %s, %s ---> %s\n", srcCode, srcDtat.location, destCode, destData.location)
		return
	}
	res = true
	d := getDomainList(srcDtat)
	srcDtat.domainlist = d
	return
}

func getDomainList(srcInfo Data) (domainlist []string) {
	var lo string

	if srcInfo.location == "us" {
		lo = "fps_us_id"
	} else {
		lo = "fps_as_id"
	}
	sqlStr := fmt.Sprintf("select domain from fps_domain where %s = ?", lo)
	rows, err := DB.Query(sqlStr, srcInfo.id)
	if err != nil {
		fmt.Printf("query failed: %v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var domain string
		err := rows.Scan(&domain)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
		}
		domainlist = append(domainlist, domain)
	}
	return
}

func inserData(masterDomain, subDomain string, srcId, destId int) {
	sqlStr := `insert into fps_domain_change(domain, sub_domain, src_fps_id, dest_fps_id, status, update_time, create_time,memo)
	values (?, ?, ?, ?, 1, now(), now(), '[added by webdevops]')`
	_, err := DB.Exec(sqlStr, masterDomain, subDomain, srcId, destId)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
}

func updataDomainRelationshIP(flag, domain, sub string, fpsId int) {
	if flag == "us" {
		sqlStr := "update fps_domain set fps_us_id = ? where domain in (? , ?)"
		_, err := DB.Exec(sqlStr, fpsId, domain, sub)
		if err != nil {
			fmt.Printf("update domain relationship failed, err:%v\n", err)
			return
		}
	} else if flag == "as" {
		sqlStr := "update fps_domain set fps_as_id = ? where sub_domain_as = ?"
		_, err := DB.Exec(sqlStr, fpsId, sub)
		if err != nil {
			fmt.Printf("update domain relationship failed, err:%v\n", err)
			return
		}
	}
}

func switchDomain(srcInfo, destInfo Data) {
	re := regexp.MustCompile(`\.`)
	if len(srcInfo.domainlist) == 0 {
		fmt.Printf("%s下未发现域名!\n", srcInfo.code)
		return
	}
	for _, d := range srcInfo.domainlist {
		data := re.Split(d, 2)
		masterDomain := data[0]
		if srcInfo.location == "us" {
			usDomain := fmt.Sprintf("%s.%s", "us", masterDomain)
			err := Aliyun.UpdateFpsDomain(masterDomain, domainName, "A", destInfo.ip, 600)
			if err != nil {
				fmt.Printf("Update %s.%s Domain Failed: %s\n", masterDomain, domainName, err)
			} else {
				fmt.Printf("Updata Domain %s.%s ------> %s\n", masterDomain, domainName, destInfo.ip)
			}
			err1 := Aliyun.UpdateFpsDomain(usDomain, domainName, "A", destInfo.ip, 600)
			if err1 != nil {
				fmt.Printf("Update %s Domain Failed: %s\n", usDomain, err)
			} else {
				fmt.Printf("Updata Domain %s.%s ------> %s\n", usDomain, domainName, destInfo.ip)
				inserData(fmt.Sprintf("%s.%s", masterDomain, domainName), fmt.Sprintf("%s.%s", usDomain, domainName), srcInfo.id, destInfo.id)
				updataDomainRelationshIP(srcInfo.location, fmt.Sprintf("%s.%s", masterDomain, domainName), fmt.Sprintf("%s.%s", usDomain, domainName), destInfo.id)
			}
		} else if srcInfo.location == "as" {
			asDomain := fmt.Sprintf("%s.%s", "as", data[0])
			err := Aliyun.UpdateFpsDomain(asDomain, domainName, "A", destInfo.ip, 600)
			if err != nil {
				fmt.Printf("Update %s.%s Domain Failed: %s\n", asDomain, domainName, err)
			} else {
				fmt.Printf("Updata Domain %s.%s ------> %s\n", asDomain, domainName, destInfo.ip)
				inserData(fmt.Sprintf("%s.%s", masterDomain, domainName), fmt.Sprintf("%s.%s", asDomain, domainName), srcInfo.id, destInfo.id)
				updataDomainRelationshIP(srcInfo.location, fmt.Sprintf("%s.%s", masterDomain, domainName), fmt.Sprintf("%s.%s", asDomain, domainName), destInfo.id)
			}
		}
	}
}
