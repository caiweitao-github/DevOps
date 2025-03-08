package main

import (
	"database/sql"
	"fmt"
	"util"
)

type DpsInfo struct {
	location, provider string
	num                int
}

type DpsNum struct {
	num int
}

var Db *sql.DB

func main() {
	str := []string{}
	CountDps, err := GetExceptionDpsNum(3)
	if err != nil {
		panic(err)
	}
	Data, err := GetExceptionDps()
	if err != nil {
		panic(err)
	}
	for _, v := range Data {
		if v.num > 4 {
			str1 := fmt.Sprintf("%s ---> %s(%d)\n", v.location, v.provider, v.num)
			str = append(str, str1)
		}
	}
	defer Db.Close()
	if len(str) == 0 {
		str = append(str, fmt.Sprintf("无地区异常, 当前异常机器共有%d台", CountDps))
	} else {
		str = append([]string{fmt.Sprintf("以下地区可能发生地区故障, 当前异常机器共有%d台", CountDps)}, str...)
	}
	util.SendMess2(str, "[dps异常通知]")
}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		panic(err)
	}
}

func GetExceptionDps() ([]DpsInfo, error) {
	sqlStr := "select location,provider,count(provider) as num from dps where status='3' and TIMESTAMPDIFF(MINUTE, dead_time, NOW()) > 10 and location != '' and not isnull(dead_time) group by location,provider order by num desc"
	DpsList := make([]DpsInfo, 0)

	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var dps DpsInfo
		err := rows.Scan(&dps.location, &dps.provider, &dps.num)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		DpsList = append(DpsList, dps)
	}
	return DpsList, nil
}

func GetExceptionDpsNum(status int) (int, error) {
	sqlStr := "select count(*) from dps where status = ?"
	var DpsDate DpsNum

	err := Db.QueryRow(sqlStr, status).Scan(&DpsDate.num)
	if err != nil {
		return 0, fmt.Errorf("query failed: %v", err)
	}
	return DpsDate.num, nil
}
