package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"util"
)

type UnstableDps struct {
	code                 string
	changeip_period, num int
}

type DpsGroup struct {
	id   int
	code string
}

type DpsInfo struct {
	code string
}

type Dps struct {
	id              int
	code            string
	changeip_period int
}

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.0")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2023-10-24")
		return
	}
	str := []string{}
	db_nodeops, err := util.ConnDb("", "", "")
	if err != nil {
		panic(fmt.Sprintf("connect nodeops failed: %s", err))
	}
	ecs := GetECSNode(db_nodeops)
	nodelist, err := GetExceptionDps(db_nodeops, ecs)
	if err != nil {
		panic(fmt.Sprintf("get exception dps failed: %s", err))
	}
	if len(nodelist) > 0 {
		db_db, err := util.ConnDb("", "", "")
		if err != nil {
			panic(fmt.Sprintf("connect db failed: %s", err))
		}
		newdpslist, err := GetNewDps(db_nodeops, len(nodelist))
		for _, i := range nodelist {
			if (i.changeip_period == 200 || i.changeip_period == 320) && i.num > 50 {
				str = append(str, CreatMess(db_db, i))
			} else if (i.changeip_period == 630 || i.changeip_period == 1860) && i.num > 15 {
				res := FixChangeIp(db_db, i, &newdpslist)
				str = append(str, res)
			} else if (i.changeip_period == 2760 || i.changeip_period == 3660) && i.num > 5 {
				res := FixChangeIp(db_db, i, &newdpslist)
				str = append(str, res)
			} else if (i.changeip_period == 10800 || i.changeip_period == 14400) && i.num > 2 {
				res := FixChangeIp(db_db, i, &newdpslist)
				str = append(str, res)
			}
		}
	} else {
		log.Println("None Unstable Dps!")
	}
	if len(str) > 0 {
		util.SendMess2(str, "[异常换IP通知]")
	}
}

func FixChangeIp(Db *sql.DB, dpsinfo UnstableDps, newdpslist *[]DpsInfo) (str string) {
	code, err := GetDps(newdpslist)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	new_data, err := GetData(Db, code)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	old_data, err := GetData(Db, dpsinfo.code)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	str = fmt.Sprintf("%s <---> %s", CreatMess(Db, dpsinfo), code)
	UpdateDpsGroup(Db, dpsinfo.changeip_period, new_data.changeip_period, old_data.id, new_data.id)
	return
}

func GetData(Db *sql.DB, code string) (Dps, error) {
	sqlStr := "select id,code,changeip_period from dps where code = ?"
	var dps Dps
	err := Db.QueryRow(sqlStr, code).Scan(&dps.id, &dps.code, &dps.changeip_period)
	if err != nil {
		return Dps{}, fmt.Errorf("%s query failed: %v", code, err)
	}
	return dps, nil
}

func GetDps(dpslist *[]DpsInfo) (string, error) {
	if len(*dpslist) == 0 {
		return "", fmt.Errorf("empty dpslist")
	}
	element := (*dpslist)[0]
	*dpslist = (*dpslist)[1:]
	return element.code, nil
}

func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
func GetExceptionDps(Db *sql.DB, ecs []string) ([]UnstableDps, error) {
	sqlStr := "select dps_code,changeip_period,count(dps_code) as num from dps_changeip_history where is_expected=0 and is_valid=1 and change_time >=(NOW() - interval 24 hour) group by dps_code,changeip_period having count(dps_code) > 1 order by count(dps_code) desc"
	DpsList := make([]UnstableDps, 0)
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var dps UnstableDps
		err := rows.Scan(&dps.code, &dps.changeip_period, &dps.num)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		ok := Find(ecs, dps.code)
		if !ok {
			DpsList = append(DpsList, dps)
		}
	}
	return DpsList, nil
}

func GetDpsGroup(Db *sql.DB, code string) ([]DpsGroup, error) {
	sqlStr := "select dpsgroup.id,dpsgroup.code from dpsgroup,dps_group_relation,dps where dps.code = ? and dps.id = dps_group_relation.dps_id and dps_group_relation.group_id = dpsgroup.id"
	GroupInfo := make([]DpsGroup, 0)
	rows, err := Db.Query(sqlStr, code)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var group DpsGroup
		err := rows.Scan(&group.id, &group.code)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		GroupInfo = append(GroupInfo, group)
	}
	return GroupInfo, nil
}

func GetNewDps(Db *sql.DB, num int) ([]DpsInfo, error) {
	sqlStr := "select dps_code from dps_devops_grade where exception_time >= CURRENT_TIMESTAMP - INTERVAL 7 DAY and is_valid=true group by  dps_code order by sum(score) desc limit ?"
	GroupInfo := make([]DpsInfo, 0)
	rows, err := Db.Query(sqlStr, num)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var dps DpsInfo
		err := rows.Scan(&dps.code)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		GroupInfo = append(GroupInfo, dps)
	}
	return GroupInfo, nil
}

func UpdateDpsGroup(Db *sql.DB, old_change_time, new_change_time, old_id, new_id int) {
	now := time.Now()
	ChangePeriod := "UPDATE dps SET changeip_period = ? WHERE id = ?"
	ChangeGroup := "UPDATE dps_group_relation SET dps_id = ?, update_time = ? WHERE dps_id = ?"
	ChangeGroup2 := "UPDATE dps_group_relation SET dps_id = ?, update_time = NOW() WHERE dps_id = ? and not update_time = ?"

	data := []struct {
		query string
		args  []interface{}
	}{
		{ChangePeriod, []interface{}{new_change_time, old_id}},
		{ChangePeriod, []interface{}{old_change_time, new_id}},
		{ChangeGroup, []interface{}{old_id, now.Format("2006-01-02 15:04:05"), new_id}},
		{ChangeGroup2, []interface{}{new_id, old_id, now.Format("2006-01-02 15:04:05")}},
	}

	for _, da := range data {
		_, err := Db.Exec(da.query, da.args...)
		if err != nil {
			panic(fmt.Sprintf("update faile: %s", err))
		}
	}
}

func CreatMess(Db *sql.DB, dps UnstableDps) (str string) {
	group_info := []string{}
	group_code, err := GetDpsGroup(Db, dps.code)
	if err != nil {
		log.Printf("%s query group err: %s", dps.code, err)
	}
	for _, x := range group_code {
		group_info = append(group_info, x.code)
	}
	format_res := strings.Join(group_info, "")
	str = fmt.Sprintf("%s(%d:%s): %d", dps.code, dps.changeip_period, format_res, dps.num)
	return
}

func GetECSNode(Db *sql.DB) (escNode []string) {
	sqlStr := "select code from dps where dps_type = 6 and status in (1, 3)"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		fmt.Printf("query failed: %v", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		err := rows.Scan(&code)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil
		}
		escNode = append(escNode, code)
	}
	return
}
