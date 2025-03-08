package main

import (
	"Aliyun"
	"database/sql"
	"fmt"
	"math/rand"
	"regexp"
	"time"
	"util"
)

const FpsTestCode = ""

const FpsBakCode = ""

var Db *sql.DB

var logInfo, logError = util.InitLog("/data/kdl/log/devops/fpsDomainAutoSwitch.log", "[INFO] ", "[ERROR] ")

var FPSIP string

var FPSBAKIP string

var FPSID int

var Tite = "[fps域名切换通知]"

type Data struct {
	OrderId, Host string
}

type Node struct {
	AsCode, UsCode string
	AsIP, UsIP     string
	AsId, UsId     int
}

type TmpData struct {
	Code, Domain string
	Id           int
}

var TTL int64 = 600

var Domain = ""

func main() {
	defer trackTime(time.Now())
	d := getOrderId()
	for _, v := range d {
		v.switchDomain()
	}

}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		logError.Fatalf("Init Db Connct Fail : %s", err)
	}
	err = getFpsTest()
	if err != nil {
		logError.Fatalf("getFPSIP Failed : %s", err)
	}
}

func trackTime(pre time.Time) {
	elapsed := time.Since(pre)
	logInfo.Print("Time Use:", elapsed)
}

func getFpsTest() (err error) {
	var ip string
	var id int
	sqlStr := "select id,login_ip from fps where code = ?"
	err = Db.QueryRow(sqlStr, FpsTestCode).Scan(&id, &ip)
	if err != nil {
		return
	} else {
		FPSIP = ip
		FPSID = id
	}
	sqlStr1 := "select login_ip from fps where code = ?"
	err = Db.QueryRow(sqlStr1, FpsBakCode).Scan(&ip)
	if err != nil {
		return
	} else {
		FPSBAKIP = ip
		return
	}

}

func getOrderId() (dataList []Data) {
	sqlStr := `select proxy_order.orderid,order_fps_conf.host from order_fps_conf,proxy_order 
	where host in (select domain from fps_domain where fps_us_id = (select id from fps where code = ?)) 
	and proxy_order.id = order_fps_conf.order_id and order_fps_conf.status = 1`
	orderExSql := `select orderid from pay_history where orderid = ? limit 1`
	rows, err := Db.Query(sqlStr, FpsTestCode)
	if err != nil {
		logError.Fatalf("Query Failed : %s", err)
	}
	defer rows.Close()
	for rows.Next() {
		var orderid, host string
		if err := rows.Scan(&orderid, &host); err != nil {
			logError.Fatalf("Scan Fail : %s", err)
		}
		var ex string
		err := Db.QueryRow(orderExSql, orderid).Scan(&ex)
		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			logError.Fatalf("Get PayHistory Failed : %s", err)
		}
		dataList = append(dataList, Data{orderid, host})
	}
	return
}

func getFpsNode() (n Node) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	asSql := "select code,login_ip,id from fps where location_code = 'as' and status = 1 and login_ip not in(?, ?)"
	usSql := "select code,login_ip,id from fps where location_code = 'us' and status = 1 and login_ip not in(?, ?)"
	rows, err := Db.Query(asSql, FPSIP, FPSBAKIP)
	asList := make([]TmpData, 0)
	usList := make([]TmpData, 0)
	if err != nil {
		logError.Fatalf("Query Failed : %s", err)
	}
	for rows.Next() {
		var code, domain string
		var id int
		if err := rows.Scan(&code, &domain, &id); err != nil {
			logError.Fatalf("Scan Fail : %s", err)
		}
		asList = append(asList, TmpData{code, domain, id})
	}
	rows, err = Db.Query(usSql, FPSIP, FPSBAKIP)
	if err != nil {
		logError.Fatalf("Query Failed : %s", err)
	}
	for rows.Next() {
		var code, domain string
		var id int
		if err := rows.Scan(&code, &domain, &id); err != nil {
			logError.Fatalf("Scan Fail : %s", err)
		}
		usList = append(usList, TmpData{code, domain, id})
	}
	usRand := rng.Intn(len(usList))
	asRand := rng.Intn(len(asList))
	n = Node{
		AsCode: asList[asRand].Code,
		UsCode: usList[usRand].Code,
		AsIP:   asList[asRand].Domain,
		UsIP:   usList[usRand].Domain,
		AsId:   asList[asRand].Id,
		UsId:   usList[usRand].Id,
	}
	return
}

func (d Data) switchDomain() {
	re := regexp.MustCompile(`\.kdlfps.com`)
	host := re.Split(d.Host, 2)[0]
	usDomain := fmt.Sprintf("%s.%s", "us", host)
	asDomain := fmt.Sprintf("%s.%s", "as", host)
	n := getFpsNode()
	err := Aliyun.UpdateFpsDomain(host, Domain, "A", n.UsIP, TTL)
	errus := Aliyun.UpdateFpsDomain(usDomain, Domain, "A", n.UsIP, TTL)
	erras := Aliyun.UpdateFpsDomain(asDomain, Domain, "A", n.AsIP, TTL)
	if err != nil || errus != nil || erras != nil {
		fmt.Printf("Update Domain Failed: %s\n", err)
		util.SendMess2([]string{fmt.Sprintf("%s Update Domain Failed.", d.OrderId)}, Tite)
		return
	}
	inserErr := inserData(fmt.Sprintf("%s.%s", host, Domain), fmt.Sprintf("%s.%s", usDomain, Domain), FPSID, n.UsId)
	inserErr1 := inserData(fmt.Sprintf("%s.%s", host, Domain), fmt.Sprintf("%s.%s", asDomain, Domain), FPSID, n.AsId)
	upDtaeErr := updataDomainRelationshIP("us", d.Host, fmt.Sprintf("%s.%s", usDomain, Domain), n.UsId)
	upDtaeErr2 := updataDomainRelationshIP("as", d.Host, fmt.Sprintf("%s.%s", asDomain, Domain), n.AsId)
	if inserErr != nil || inserErr1 != nil || upDtaeErr != nil || upDtaeErr2 != nil {
		util.SendMess2([]string{fmt.Sprintf("%s Sync Db Failed.", d.OrderId)}, Tite)
		return
	}
	util.SendMess2([]string{fmt.Sprintf("%s\n%s ---> %s(%s)\n%s ---> %s(%s)\n%s ---> %s(%s)", d.OrderId, host, n.UsCode, n.UsIP, usDomain, n.UsCode, n.UsIP, asDomain, n.AsCode, n.AsIP)}, Tite)
}

func inserData(masterDomain, subDomain string, srcId, destId int) error {
	sqlStr := `insert into fps_domain_change(domain, sub_domain, src_fps_id, dest_fps_id, status, update_time, create_time,memo)
	values (?, ?, ?, ?, 1, now(), now(), '[added by FpsDomainAutoSwitch]')`
	_, err := Db.Exec(sqlStr, masterDomain, subDomain, srcId, destId)
	if err != nil {
		return err
	}
	return nil
}

func updataDomainRelationshIP(flag, domain, sub string, fpsId int) error {
	if flag == "us" {
		sqlStr := "update fps_domain set fps_us_id = ? where domain in (? , ?)"
		_, err := Db.Exec(sqlStr, fpsId, domain, sub)
		if err != nil {
			return err
		}
	} else if flag == "as" {
		sqlStr := "update fps_domain set fps_as_id = ? where sub_domain_as = ?"
		_, err := Db.Exec(sqlStr, fpsId, sub)
		if err != nil {
			return err
		}
	}
	return nil
}
