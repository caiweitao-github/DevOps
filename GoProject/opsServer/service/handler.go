package service

import "fmt"

type StatusHandler struct {
	updateDbFunc map[string]func(int, string)
}

func NewUpdateStatusHandler() *StatusHandler {
	return &StatusHandler{
		updateDbFunc: map[string]func(int, string){
			"TFS":  updateTfsStatus,
			"CAB":  updateCabStatus,
			"SFPS": updateSfpsStatus,
		},
	}
}

func (s *StatusHandler) UpdateStatus(category string, status int, code string) {
	if updateFunc, exists := s.updateDbFunc[category]; exists {
		updateFunc(status, code)
	}
}

func updateSfpsStatus(status int, code string) {
	sqlStr := "update sfps set status = ? where code = ?"
	db.Exec(sqlStr, status, code)
}

func updateTfsStatus(status int, code string) {
	sqlStr := "update transfer_server set status = ? where code = ?"
	Kdlnode.Exec(sqlStr, status, code)
}

func updateCabStatus(status int, code string) {
	sqlStr := "update gateway_cab set status = ? where code = ?"
	Flnode.Exec(sqlStr, status, code)
}

func updateNodeDownHistory(code string) (err error) {
	var id int
	var recoverTime int
	sqlStr := `select id, IF(recover_time IS NULL, 1, 0) from node_down_history where code = ? order by down_time desc limit 1`
	err = KdlStat.QueryRow(sqlStr, code).Scan(&id, &recoverTime)
	if err != nil {
		return
	}
	if recoverTime == 1 {
		sqlStr := `update node_down_history set recover_time=now(),is_valid=? where id=?`
		KdlStat.Exec(sqlStr, 1, id)
	}
	return
}

func inserData(code, ip, provider string) {
	level, err := getOrderLevel(code)
	if err != nil {
		return
	}
	sqlStr := `insert into node_down_history(code,ip,level,provider,is_valid,down_time) values(?,?,?,?,0,now())`
	KdlStat.Exec(sqlStr, code, ip, level, provider)
}

func getOrderLevel(code string) (res string, err error) {
	var level int
	sqlStr := `select sfps_type from sfps where code = ?`
	err = db.QueryRow(sqlStr, code).Scan(&level)
	if err != nil {
		fmt.Println(err)
		return
	}
	if level == 1 {
		res = "21"
	} else if level == 2 {
		res = "22"
	}
	return
}

func timeConversion(ti int64) string {
	if ti < 60 {
		return fmt.Sprintf("%d秒", ti)
	} else {
		return fmt.Sprintf("%d分钟", ti/60)
	}
}

func getNodeProvider(code string) (res string) {
	sqlStr := "select provider from sfps where code = ?"
	err := db.QueryRow(sqlStr, code).Scan(&res)
	if err != nil {
		res = "None"
	}
	return
}
