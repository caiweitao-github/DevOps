package service

import (
	"Aliyun"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

type ServiceConfig struct {
	Protocol string
	Address  string
}

type KDLService interface {
	GetData(parameter Parameter, result *[]Result) error
	CheckStatus(parameter NodeList, _ *struct{}) error
	GetNode(_ struct{}, result *[]string) error
	GetOrderData(_ struct{}, result *[]OrderDtat) error
	GetSfpsData(_ struct{}, result *[]SfpsData) error
	CheckSfpsStatus(parameter NodeList, _ *struct{}) error
	QueryOrderID(secretid string, result *string) error
	ReportForbidIP(data NginxForbidData, _ *struct{}) error
	UnForbidIP(reportip ReportIP, ip *[]string) error
	UpdataTpsDomainRecord(data DomainRecordDate, _ *struct{}) error
	UpdateCache(data JipFDate, _ *struct{}) error
	ModifyTdpsStatus(parameter DataEntry, _ *struct{}) error
	ReportIpPool(data IpPool, _ *struct{}) error
	GetCabNode(_ struct{}, result *[]string) error
	GetCabLine(parameter Parameter, result *[]Result) error
}

type KDL struct{}

type Parameter struct {
	ServerIP string `form:"server_ip" json:"server_ip" binding:"required"`
	Num      int    `form:"num" json:"num" binding:"required"`
}

type Result struct {
	ServerCode string `json:"server_code"`
	NodeIP     string `json:"ip"`
	NodePort   string `json:"port"`
}

type OrderDtat struct {
	OrderID  string  `json:"orderid"`
	SecretID *string `json:"secret_id"`
}

type DataEntry struct {
	IP   string `json:"ip"`
	Code string `json:"code"`
	Log  string `json:"log"`
}

type NodeList struct {
	Category  string      `json:"category"`
	AliveList []DataEntry `json:"alive_list"`
	DeadList  []DataEntry `json:"dead_list"`
}

type SfpsData struct {
	ID        int     `json:"id"`
	Code      string  `json:"code"`
	IP        string  `json:"ip"`
	Port      string  `json:"port"`
	SocksPort string  `json:"socks_port"`
	UserName  *string `json:"username"`
	PassWord  *string `json:"password"`
}

type NginxForbidData struct {
	Code string `json:"code"`
	IP   string `json:"ip"`
	Data NginxData
}

type NginxData struct {
	Count      string `json:"count"`
	ForbidTime string `json:"forbid_time"`
	Reason     string `json:"reason"`
	OrderID    string `json:"orderid"`
}

type ReportIP struct {
	Code   string   `json:"code"`
	IpList []string `json:"ip_list"`
}

type DomainRecordDate struct {
	KeyWord    string `json:"keyword"`
	DomainName string `json:"domain"`
	Value      string `json:"ip"`
}

type JipF struct {
	Country    string `json:"country_code"`
	ProviderID string `json:"provider_id"`
}

type JipFCount struct {
	Country string `json:"country_code"`
	Count   int    `json:"count"`
}

type JipFDate struct {
	JipFList      []JipF      `json:"jipf_list"`
	JipFCountList []JipFCount `json:"jipf_count_list"`
}

type Pool struct {
	Period string `json:"period"`
	Num    int    `json:"num"`
}

type IpPool struct {
	Flag string `json:"flag"`
	Data []Pool `json:"data"`
}

var (
	FailedServer    = make(map[string]int)
	Kdlnode         *sql.DB
	db              *sql.DB
	KdlStat         *sql.DB
	Flnode          *sql.DB
	rdb             *redis.Client
	rdbcrs          *redis.Client
	iprdb           *redis.Client
	ctx             = context.Background()
	forbidipKey     = "_nginx_forbidip"
	unforbidipKey   = "_nginx_unforbidip"
	notifySfpsUrl   = ""
	notifySfpsTitle = "[海外代理静态型异常通知]"
	notifyUrl       = ""
	notifyTitle     = "[中转服务异常告警]"
)

const (
	UP              = 1
	DOWN            = 3
	Pause           = 2
	FailedCount     = 10
	SfpsFailedCount = 6
	ServiceName     = "KDL"
	FailedList      = "FailedNodeList"
	SfpsDownKey     = "SfpsDownMap"
	SfpsDownHash    = "fps_master:sfps_down_hash"
	SfpsRecoverHash = "fps_master:sfps_recover_hash"
)

var OrderStatus = []interface{}{"TRADE_SUCCESS", "OWING", "CLOSE_WAIT_CHARGE"}

var RpcConfig, GinConfig ServiceConfig

func init() {
	RpcConfig = ServiceConfig{
		Protocol: "tcp",
		Address:  "127.0.0.1:8098",
	}
	GinConfig = ServiceConfig{
		Protocol: "",
		Address:  "10.0.6.130:8099",
	}
	var err error
	rdb, err = util.RedisDB()
	if err != nil {
		panic(err)
	}
	iprdb, err = util.IPPoolRedisDB()
	if err != nil {
		panic(err)
	}
	rdbcrs, err = util.RedisCrs()
	if err != nil {
		panic(err)
	}
	Kdlnode, err = util.KdlnodeDB()
	if err != nil {
		panic(err)
	}
	db, err = util.dbDB()
	if err != nil {
		panic(err)
	}
	KdlStat, err = util.KdlStat()
	if err != nil {
		panic(err)
	}
	Flnode, err = util.FlNode()
	if err != nil {
		panic(err)
	}
}

func (g *KDL) GetNode(_ struct{}, result *[]string) error {
	sqlStr := "select login_ip from transfer_server where status in (1,2)"
	rows, err := Kdlnode.Query(sqlStr)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return err
		}
		*result = append(*result, s)
	}
	return nil
}

func (g *KDL) GetData(parameter Parameter, result *[]Result) error {
	sqlStr := `SELECT transfer_server.code,ip,port FROM node,transfer_server
	WHERE RAND() <= 1 and transfer_server.status in (1,2) and transfer_server.login_ip = node.ip
	and transfer_server.login_ip = ? and source in (select source from node where ip = ? group by source) and node.status in (1,2) ORDER BY RAND()
	LIMIT ?`
	rows, err := Kdlnode.Query(sqlStr, parameter.ServerIP, parameter.ServerIP, parameter.Num)
	if err != nil {
		return errors.New("get data err")
	}
	defer rows.Close()
	for rows.Next() {
		var r Result
		err := rows.Scan(&r.ServerCode, &r.NodeIP, &r.NodePort)
		if err != nil {
			return errors.New("building data err")
		}
		*result = append(*result, r)
	}
	return nil
}

func (g *KDL) CheckStatus(parameter NodeList, _ *struct{}) error {
	for _, node := range parameter.AliveList {
		v, ex := FailedServer[node.Code]
		if ex {
			if v >= FailedCount {
				handler := NewUpdateStatusHandler()
				handler.UpdateStatus(parameter.Category, UP, node.Code)
				util.StaffFeiShuNotify(notifyUrl, notifyTitle, []string{fmt.Sprintf("%s %s recover!", node.Code, node.IP)})
			}
			delete(FailedServer, node.Code)
		}
	}

	for _, node := range parameter.DeadList {
		FailedServer[node.Code] = FailedServer[node.Code] + 1
		if v := FailedServer[node.Code]; v == FailedCount {
			handler := NewUpdateStatusHandler()
			handler.UpdateStatus(parameter.Category, DOWN, node.Code)
			util.StaffFeiShuNotify(notifyUrl, notifyTitle, []string{fmt.Sprintf("%s %s down, Error Log: %s", node.Code, node.IP, node.Log)})
			util.Notify()
		}
	}
	return nil
}

func (g *KDL) GetOrderData(_ struct{}, result *[]OrderDtat) error {
	sqlStr := `select orderid, secret_id from proxy_order where end_time > now() or (end_time is null and status in (?, ?, ?))`
	rows, err := db.Query(sqlStr, OrderStatus...)
	if err != nil {
		return errors.New("get data err")
	}
	defer rows.Close()
	for rows.Next() {
		var r OrderDtat
		err := rows.Scan(&r.OrderID, &r.SecretID)
		if err != nil {
			return errors.New("building data err")
		}
		*result = append(*result, r)
	}
	return nil
}

func (g *KDL) GetSfpsData(_ struct{}, result *[]SfpsData) error {
	sqlStr := `select id,code,ip,port,socks_port,username,password from sfps where status in (1, 3)`
	rows, err := db.Query(sqlStr)
	if err != nil {
		return errors.New("get data err")
	}
	defer rows.Close()
	for rows.Next() {
		var r SfpsData
		err := rows.Scan(&r.ID, &r.Code, &r.IP, &r.Port, &r.SocksPort, &r.UserName, &r.PassWord)
		if err != nil {
			return errors.New("building data err")
		}
		*result = append(*result, r)
	}
	return nil
}

func (g *KDL) CheckSfpsStatus(parameter NodeList, _ *struct{}) error {
	downMess := []string{}
	recoverMess := []string{}
	for _, node := range parameter.AliveList {
		if v := rdb.HExists(ctx, SfpsDownKey, node.Code).Val(); v {
			num, _ := strconv.Atoi(rdb.HGet(ctx, SfpsDownKey, node.Code).Val())
			if num >= SfpsFailedCount {
				provider := getNodeProvider(node.Code)
				now := time.Now().Unix()
				downTimeStr := rdbcrs.HGet(ctx, SfpsDownHash, node.Code).Val()
				downTimeInt, _ := strconv.ParseInt(downTimeStr, 10, 64)
				newTime := timeConversion(now - downTimeInt)
				recoverMess = append(recoverMess, fmt.Sprintf("%s %s %s recover! (%s)", node.Code, node.IP, provider, newTime))
				handler := NewUpdateStatusHandler()
				handler.UpdateStatus(parameter.Category, UP, node.Code)
				rdbcrs.HSet(ctx, SfpsRecoverHash, node.Code, now)
				updateNodeDownHistory(node.Code)
			}
			rdb.HDel(ctx, SfpsDownKey, node.Code)
		}
	}
	if len(recoverMess) > 0 {
		util.FeiShuNotify(notifySfpsUrl, notifySfpsTitle, recoverMess)
	}
	for _, node := range parameter.DeadList {
		if res := rdb.HSetNX(ctx, SfpsDownKey, node.Code, 1).Val(); !res {
			rdb.HIncrBy(ctx, SfpsDownKey, node.Code, 1)
		}
		if v, _ := strconv.Atoi(rdb.HGet(ctx, SfpsDownKey, node.Code).Val()); v == SfpsFailedCount {
			provider := getNodeProvider(node.Code)
			downMess = append(downMess, fmt.Sprintf("%s %s %s down!", node.Code, node.IP, provider))
			handler := NewUpdateStatusHandler()
			handler.UpdateStatus(parameter.Category, DOWN, node.Code)
			rdbcrs.HSet(ctx, SfpsDownHash, node.Code, int(time.Now().Unix()))
			inserData(node.Code, node.IP, provider)
		}
	}
	if len(downMess) > 0 {
		util.FeiShuNotify(notifySfpsUrl, notifySfpsTitle, downMess)
	}
	return nil
}

func (g *KDL) QueryOrderID(secretid string, result *string) error {
	sqlStr := `select orderid from proxy_order where secret_id = ?`
	err := db.QueryRow(sqlStr, secretid).Scan(result)
	if err != nil {
		return errors.New("get data err")
	}
	return nil
}

func (g *KDL) UnForbidIP(reportip ReportIP, ip *[]string) error {
	for _, v := range reportip.IpList {
		rdbcrs.HDel(ctx, reportip.Code+forbidipKey, v)
		rdbcrs.SRem(ctx, reportip.Code+unforbidipKey, v)
	}
	r := rdbcrs.SMembers(ctx, reportip.Code+unforbidipKey).Val()
	rdbcrs.Del(ctx, reportip.Code+unforbidipKey)
	*ip = append(*ip, r...)
	return nil
}

func (g *KDL) ReportForbidIP(data NginxForbidData, _ *struct{}) error {
	v, err := json.Marshal(data.Data)
	if err != nil {
		return errors.New("check data err")
	}
	rdbcrs.HSet(ctx, data.Code+forbidipKey, data.IP, v)
	return nil
}

func (g *KDL) UpdataTpsDomainRecord(data DomainRecordDate, _ *struct{}) error {
	err := Aliyun.UpdateTpsDomain(data.KeyWord, data.DomainName, "A", data.Value, 1)
	if err != nil {
		return fmt.Errorf("%v", err.Error())
	}

	return nil
}

func (g *KDL) UpdateCache(data JipFDate, _ *struct{}) error {
	t := time.Now().Format(time.DateTime)
	t1 := time.Now().Format(time.DateOnly)
	for _, d := range data.JipFList {
		rdbcrs.SAdd(ctx, fmt.Sprintf("JipF-%s", t1), fmt.Sprintf("%s:%s", d.Country, d.ProviderID))
	}
	for _, d := range data.JipFCountList {
		rdbcrs.SAdd(ctx, fmt.Sprintf("JipFCount-%s", t), fmt.Sprintf("%s:%d", d.Country, d.Count))
	}
	return nil
}

func (g *KDL) ModifyTdpsStatus(parameter DataEntry, _ *struct{}) error {
	// if updateFunc, found := statusUpdateFuncs[parameter.Category]; found {
	// 	updateFunc(UP, node.Code)
	// }
	return nil
}

func (g *KDL) ReportIpPool(data IpPool, _ *struct{}) error {
	yesterday := time.Now().Add(-24 * time.Hour).Format(time.DateOnly)
	key := fmt.Sprintf("%s:%s", data.Flag, yesterday)
	for _, ip := range data.Data {
		iprdb.HSet(ctx, key, ip.Period, ip.Num)
	}
	iprdb.Expire(ctx, key, 24*60*60*90*time.Second)
	return nil
}

func (g *KDL) GetCabNode(_ struct{}, result *[]string) error {
	sqlStr := "select ip from gateway_cab where status in (1, 3)"
	rows, err := Flnode.Query(sqlStr)
	if err != nil {
		return errors.New("get data err")
	}
	defer rows.Close()
	for rows.Next() {
		var ip string
		err := rows.Scan(&ip)
		if err != nil {
			return errors.New("building data err")
		}
		*result = append(*result, ip)
	}
	return nil
}

func (g *KDL) GetCabLine(parameter Parameter, result *[]Result) error {
	var code string
	var randNum int
	randSql := "select FLOOR(RAND() * COUNT(*)) from cn_node where status = 1 and proxy_ip = ?"
	err := Flnode.QueryRow(randSql, parameter.ServerIP).Scan(&randNum)
	if err != nil {
		return errors.New("get randnum err")
	}
	serverCodeSql := "select code from gateway_cab where ip = ? and status  in (1, 3)"
	err = Flnode.QueryRow(serverCodeSql, parameter.ServerIP).Scan(&code)
	if err != nil {
		return errors.New("get server code err")
	}
	sqlStr := `SELECT t0.proxy_ip, t0.proxy_port FROM ( SELECT * FROM flnode.cn_node t WHERE t.proxy_ip = ? AND CAST(t.status AS SIGNED) = 1 ) t0 
	INNER JOIN ( SELECT t1.source FROM flnode.cn_node t1 WHERE t1.proxy_ip = ? GROUP BY t1.source) t4 ON t0.source = t4.source LIMIT ? OFFSET ?`
	rows, err := Flnode.Query(sqlStr, parameter.ServerIP, parameter.ServerIP, parameter.Num, randNum)
	if err != nil {
		return errors.New("get data err")
	}
	defer rows.Close()
	for rows.Next() {
		var r Result
		err := rows.Scan(&r.NodeIP, &r.NodePort)
		if err != nil {
			return errors.New("building data err")
		}
		r.ServerCode = code
		*result = append(*result, r)
	}
	return nil
}
