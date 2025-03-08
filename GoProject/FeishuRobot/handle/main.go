package handle

import (
	"Aliyun"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"util"

	"github.com/pkg/errors"
)

var (
	Db       *sql.DB
	logFile  = "/data/kdl/log/opsServer/feishuRobot.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

type tps struct {
	srcID    int
	destID   int
	srcCode  string
	destCode string
	ip       string
	domain   string
	domain2  string
}

type node struct {
	Code       string
	OrderID    string
	EndTime    string
	ExpireTime string
	DiffTime   int
}

type sfpsnode struct {
	id   int
	code string
}

type kpsnode struct {
	id    int
	level int
	code  string
}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		logError.Printf("init db err: %v", err)
	}
}

func (S sfpsnode) updateSfpsNodeStatus() string {
	sqlStr := "update sfps set status = 4 where code = ?"
	Db.Exec(sqlStr, S.code)
	return S.code + "到期下架."
}

func (S kpsnode) updateKpsNodeStatus() string {
	sqlStr := "update kps set status = 4 where code = ?"
	Db.Exec(sqlStr, S.code)
	return S.code + "到期下架."
}

func (S sfpsnode) checkSfpsNodeIsExpired() (r string) {
	sqlStr := `select sfps.code from sfps,order_sfps,proxy_order where proxy_order.level in ('21', '22') and 
	proxy_order.status in ('TRADE_SUCCESS', 'EXPIRED') and proxy_order.id = order_sfps.order_id and order_sfps.sfps_id = sfps.id and 
	proxy_order.end_time < NOW() and  DATE_FORMAT(proxy_order.end_time, '%Y-%m-%d %H:%i') <= DATE_FORMAT(sfps.expire_time, '%Y-%m-%d %H:%i')  and sfps.id = ?`
	var code string
	err := Db.QueryRow(sqlStr, S.id).Scan(&code)
	if err == sql.ErrNoRows {
		countOrd := getOrderCount("order_sfps", "sfps_id", S.id)
		if countOrd {
			r = S.updateSfpsNodeStatus()
		}
	} else if err == nil {
		r = S.updateSfpsNodeStatus()
	}
	return
}

func (S kpsnode) checkKpsNodeIsExpired() (r string) {
	sqlStr := `select kps.code from kps,order_kps,proxy_order where proxy_order.level in ('64', '66', '6') and 
	proxy_order.status in ('TRADE_SUCCESS', 'EXPIRED') and proxy_order.id = order_kps.order_id and order_kps.kps_id = kps.id and 
	proxy_order.end_time < NOW() and  DATE_FORMAT(proxy_order.end_time, '%Y-%m-%d %H:%i') <= DATE_FORMAT(kps.expire_time, '%Y-%m-%d %H:%i')  and kps.id = ?`
	var code string
	err := Db.QueryRow(sqlStr, S.id).Scan(&code)
	if err == sql.ErrNoRows {
		countOrd := getOrderCount("order_kps", "kps_id", S.id)
		if countOrd {
			r = S.updateKpsNodeStatus()
		}
	} else if err == nil {
		r = S.updateKpsNodeStatus()
	}
	return
}

func CheckKpsNode() (res string, e error) {
	nodeList, err := getKpsData()
	if err != nil {
		e = err
		return
	}
	if len(nodeList) < 1 {
		e = errors.New("没有符合条件的节点")
		return
	}
	var r = make([]string, 0, 50)
	for _, i := range nodeList {
		if i.level != 5 {
			r = append(r, fmt.Sprintf("%s为非独享机器, 请人工检查.", i.code))
		} else {
			sqlStr := `select kps.code,proxy_order.orderid,proxy_order.end_time,kps.expire_time,
			TIMESTAMPDIFF(HOUR, kps.expire_time, proxy_order.end_time) as diff from kps,order_kps,
			proxy_order where proxy_order.level in ('64', '66', '6') and proxy_order.status = 'TRADE_SUCCESS' and proxy_order.id = order_kps.order_id and
			order_kps.kps_id = kps.id and proxy_order.end_time > NOW() and TIMESTAMPDIFF(HOUR, kps.expire_time, proxy_order.end_time) >= 1 and kps.id = ?`
			var s node
			err := Db.QueryRow(sqlStr, i.id).Scan(&s.Code, &s.OrderID, &s.EndTime, &s.ExpireTime, &s.DiffTime)
			if err == sql.ErrNoRows {
				res := i.checkKpsNodeIsExpired()
				if res != "" {
					r = append(r, res)
				}
			} else if err != nil {
				e = err
				return
			} else {
				r = append(r, fmt.Sprintf("%s 订单: %s %s | %s, 时间差值: %d(小时)", s.Code, s.OrderID, s.EndTime, s.ExpireTime, s.DiffTime))
			}
		}
	}
	if len(r) == 0 {
		res = fmt.Sprintf("没有需要自动处理的节点, 现存异常节点数量 ➡️ %d", len(nodeList))
	} else {
		res = strings.Join(r, "\n")
	}
	return
}

func CheckSfpsNode() (res string, e error) {
	nodeList, err := getSfpsData()
	if err != nil {
		e = err
		return
	}
	if len(nodeList) < 1 {
		e = errors.New("没有符合条件的节点")
		return
	}
	var r = make([]string, 0, 50)
	for _, i := range nodeList {
		sqlStr := `select sfps.code,proxy_order.orderid,proxy_order.end_time,sfps.expire_time,TIMESTAMPDIFF(HOUR, sfps.expire_time, proxy_order.end_time) as diff from 
		sfps,order_sfps,proxy_order where proxy_order.level in ('21', '22') and proxy_order.status = 'TRADE_SUCCESS' and proxy_order.id = order_sfps.order_id and
		order_sfps.sfps_id = sfps.id and proxy_order.end_time > NOW() and TIMESTAMPDIFF(HOUR, sfps.expire_time, proxy_order.end_time) >= 1 and sfps.id = ?`
		var s node
		err := Db.QueryRow(sqlStr, i.id).Scan(&s.Code, &s.OrderID, &s.EndTime, &s.ExpireTime, &s.DiffTime)
		if err == sql.ErrNoRows {
			res := i.checkSfpsNodeIsExpired()
			if res != "" {
				r = append(r, res)
			}
		} else if err != nil {
			e = err
			return
		} else {
			r = append(r, fmt.Sprintf("%s 订单: %s %s | %s, 时间差值: %d(小时)", s.Code, s.OrderID, s.EndTime, s.ExpireTime, s.DiffTime))
		}
	}
	if len(r) == 0 {
		res = fmt.Sprintf("没有需要自动处理的节点, 现存异常节点数量 ➡️ %d", len(nodeList))
	} else {
		res = strings.Join(r, "\n")
	}
	return
}

func getSfpsData() (nodeList []sfpsnode, e error) {
	sqlStr := "select id,code from sfps where status = 3"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		e = err
		return
	}
	defer rows.Close()
	for rows.Next() {
		var s sfpsnode
		err := rows.Scan(&s.id, &s.code)
		if err != nil {
			e = err
			return
		}
		nodeList = append(nodeList, s)
	}
	return
}

func getKpsData() (nodeList []kpsnode, e error) {
	sqlStr := "select id,level,code from kps where status = 3"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		e = err
		return
	}
	defer rows.Close()
	for rows.Next() {
		var s kpsnode
		err := rows.Scan(&s.id, &s.level, &s.code)
		if err != nil {
			e = err
			return
		}
		nodeList = append(nodeList, s)
	}
	return
}

func getOrderCount(nodeCategory, nodeType string, nodeID int) (r bool) {
	sqlStr := fmt.Sprintf("select count(*) from %s where %s = ?", nodeCategory, nodeType)
	var count int
	err := Db.QueryRow(sqlStr, nodeID).Scan(&count)
	if err != nil {
		logError.Printf("getOrderCount err %v", err)
		return
	}
	if count == 0 {
		r = true
	}
	return
}

func UpdateTpsDomainRecord(tid, tpsCode string) (string, error) {
	var t tps
	sqlStr := `select tps.code,tps_domain.tps_id,host,host2 from tunnel,tps_domain,tps where tunnel.tid = ? and 
	tunnel.host = tps_domain.domain and tunnel.host2 = tps_domain.domain2 and tps_domain.tps_id = tps.id`
	err := Db.QueryRow(sqlStr, tid).Scan(&t.srcCode, &t.srcID, &t.domain, &t.domain2)
	sqlStr2 := `select id,code,login_ip from tps where code = ?`
	err1 := Db.QueryRow(sqlStr2, tpsCode).Scan(&t.destID, &t.destCode, &t.ip)
	if err != nil || err1 != nil {
		return "", errors.New("get tps data err")
	}
	err3 := t.updateDomain()
	if err3 != nil {
		return "", errors.New("切换域名失败")
	}
	return fmt.Sprintf("切换域名成功: %s ---> %s(%s), %s ---> %s(%s)", t.domain, t.ip, t.destCode, t.domain2, t.ip, t.destCode), nil
}

func (t tps) updateDomain() error {
	re := regexp.MustCompile(`\.`)
	d1 := re.Split(t.domain, 2)
	d2 := re.Split(t.domain2, 2)
	err := Aliyun.UpdateTpsDomain(d1[0], d1[1], "A", t.ip, 600)
	err1 := Aliyun.UpdateTpsDomain(d2[0], d2[1], "A", t.ip, 600)
	if err != nil || err1 != nil {
		return errors.New("切换域名失败.")
	}
	t.updateDB()
	t.insertData()
	return nil
}

func (t tps) updateDB() {
	sqlStr := "update tps_domain set tps_id = ? where domain = ? or domain2 = ?"
	Db.Exec(sqlStr, t.destID, t.domain, t.domain)
}

func (t tps) insertData() {
	sqlStr := `insert into tps_domain_change(src_domain, src_tps_id, dest_tps_id, status, update_time, create_time, memo)
	values (?, ?, ?, 1, now(), now(), '[added by FeishuRobot]'), (?, ?, ?, 1, now(), now(), '[added by FeishuRobot]')`
	Db.Exec(sqlStr, t.domain, t.srcID, t.destID, t.domain2, t.srcID, t.destID)
}
