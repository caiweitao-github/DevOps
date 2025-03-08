package main

import (
	"Aliyun"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"time"
	"util"
)

var FpsHA = "fpsbak"

var (
	logFile  = "/data/kdl/log/devops/fps_domain_check.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

const (
	domainNum     = 5
	fpsDomain     = ""
	fpsDomainLine = "" //oversea: 解析为境外，default：解析线路为默认
)

type fps struct {
	id                       int
	code, ip                 string
	avg_bandwidth, bandwidth int
}

var Db *sql.DB

var FpsTest = "fps02"

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.6")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2023-12-15")
		return
	}
	start := time.Now()
	logInfo.Println("Run Fps_Domain_Check ...")
	CheckFpsDomainStatus()
	CheckFpsDomain()
	CheckFpsTestDomain()
	exceptionFps := SyncDomainRemark()
	if len(exceptionFps) > 0 {
		util.SendMess2(exceptionFps, "[域名解析异常]")
	}
	end := time.Now()
	logInfo.Printf("Time Use: %v", end.Sub(start))
}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		logError.Printf("ConnDb Failed : %s", err)
	}
}

func CheckFpsDomainStatus() {
	sqlStr := []string{
		`UPDATE fps_domain
	JOIN (
		SELECT fps_domain.domain
		FROM fps_domain
		LEFT JOIN order_fps_conf ON order_fps_conf.host = fps_domain.domain
		WHERE order_fps_conf.status = 1 AND fps_domain.status = 1
	) AS subquery
	ON fps_domain.domain = subquery.domain
	SET fps_domain.status = 2`,

		`UPDATE fps_domain
	JOIN (
		SELECT fps_domain.domain
		FROM fps_domain
		LEFT JOIN order_fps_conf ON order_fps_conf.host = fps_domain.domain
		WHERE fps_domain.status = 2 AND fps_domain.domain NOT IN (SELECT host FROM order_fps_conf WHERE status = 1)
	) AS subquery
	ON fps_domain.domain = subquery.domain
	SET fps_domain.status = 1`}
	for _, v := range sqlStr {
		_, err := Db.Exec(v)
		if err != nil {
			logError.Printf("update fps_domain failed: %v", err)
			return
		}
	}
}

func GetFpsNode(location string) (fps, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	sqlStr := `select ANY_VALUE(fps.id) as id,ANY_VALUE(fps.code) as code,fps.login_ip as ip,
	ANY_VALUE(format(fps.avg_bandwidth/1024/128,0)) as avg_bandwidth,ANY_VALUE(fps.bandwidth) as
	bandwidth from fps left join fps_domain on fps.id = fps_domain.fps_as_id where fps.code not in (?, ?) and fps.status = 1 and fps.location_code = ? group by ip;`
	rows, err := Db.Query(sqlStr, FpsHA, FpsTest, location)
	if err != nil {
		logError.Printf("query failed: %s", err)
		return fps{}, err
	}
	defer rows.Close()
	fpsNodeList := make([]fps, 0)
	dataList := make([]fps, 0)
	for rows.Next() {
		var fpsNode fps
		err := rows.Scan(&fpsNode.id, &fpsNode.code, &fpsNode.ip, &fpsNode.avg_bandwidth, &fpsNode.bandwidth)
		if err != nil {
			logError.Printf("query fps failed: %v", err)
			return fps{}, err
		} else {
			fpsNodeList = append(fpsNodeList, fpsNode)
		}
	}
	for _, v := range fpsNodeList {
		if v.avg_bandwidth < int(float64(v.bandwidth)*0.8) {
			dataList = append(dataList, v)
		}
	}
	if len(dataList) > 0 {
		randomIndex := rng.Intn(len(dataList))
		return dataList[randomIndex], nil
	} else {
		return fpsNodeList[rng.Intn(len(fpsNodeList))], nil
	}
}

func CheckFpsDomain() {
	fpsmess := []string{}
	sqlStr := "select count(*) as num from fps_domain where status = 1 and fps_us_id != (select id from fps where code = ?)"
	rows, err := Db.Query(sqlStr, FpsTest)
	if err != nil {
		logError.Printf("query failed: %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var num int
		err := rows.Scan(&num)
		if err != nil {
			logError.Printf("query fps_domain failed: %v", err)
			return
		}
		if num < domainNum {
			fpsAsNode, err := GetFpsNode("as")
			fpsUsNode, err2 := GetFpsNode("us")
			if err != nil || err2 != nil {
				logError.Printf("Get Fps Node Err: %s", err)
				return
			}
			str, err := CreateFpsDomain(fpsAsNode, fpsUsNode)
			if err != nil {
				logError.Printf("Create Fps Domain Err: %s", err)
				return
			} else {
				fpsmess = append(fpsmess, str...)
			}
		}
	}
	if len(fpsmess) > 0 {
		util.SendMess2(fpsmess, "[新建FPS域名解析通知]")
	}
}

func CheckFpsTestDomain() {
	fpsmess := []string{}
	sqlStr := "select count(*) as num from fps_domain,fps where fps_domain.fps_us_id = fps.id and fps_domain.status = 1 and fps.code = ?"
	var num int
	err := Db.QueryRow(sqlStr, FpsTest).Scan(&num)
	if err != nil {
		logError.Printf("query failed: %s", err)
		return
	}
	if num < domainNum {
		fpsNode, err := getFpsTest()
		if err != nil {
			logError.Printf("Get FpsTest Node Err: %s", err)
			return
		}
		str, err := CreateFpsDomain(fpsNode, fpsNode)
		if err != nil {
			logError.Printf("Create FpsTest Domain Err: %s", err)
			return
		} else {
			fpsmess = append(fpsmess, str...)
		}
	}
	if len(fpsmess) > 0 {
		util.SendMess2(fpsmess, "[新建FpsTest域名解析通知]")
	}
}

func getFpsTest() (f fps, err error) {
	sqlStr := `select id,code,login_ip,format(fps.avg_bandwidth/1024/128,0) as avg_bandwidth,bandwidth 
	from fps where code = ?`
	err = Db.QueryRow(sqlStr, FpsTest).Scan(&f.id, &f.code, &f.ip, &f.avg_bandwidth, &f.bandwidth)
	return
}

func CreateFpsDomain(fpsAsNode, fpsUsNode fps) ([]string, error) {
	str := []string{}
	sqlStr := `insert into fps_domain(domain, sub_domain_as, sub_domain_us, fps_as_id, fps_us_id, status, update_time, create_time, memo)
	values(?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), "")`
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := 1
LOOP:
	for i <= 10 {
		prefix := 'a' + rng.Intn(26)
		suffix := rng.Intn(901) + 100
		domainList := map[string][]string{
			"domain":   {fmt.Sprintf("%c%d", prefix, suffix), fpsUsNode.ip, fpsAsNode.code},
			"domainAs": {fmt.Sprintf("as.%c%d", prefix, suffix), fpsAsNode.ip, fpsAsNode.code},
			"domainUs": {fmt.Sprintf("us.%c%d", prefix, suffix), fpsUsNode.ip, fpsUsNode.code},
		}
		_, err := Aliyun.QueryDomainRecord(domainList["domain"][0], fpsDomain)
		if err == nil {
			logError.Printf("%s.%s Domain Already Exist, Skip!", domainList["domain"][0], fpsDomain)
			goto LOOP
		} else {
			for _, v := range domainList {
				err := Aliyun.AddDomainRecord(v[0], fpsDomain, "A", v[1], fpsDomainLine, 600)
				if err != nil {
					logError.Printf("Create %s Domain Failed, %s", v, err)
					return nil, err
				} else {
					str = append(str, fmt.Sprintf("%s.%s ------> %s(%s)", v[0], fpsDomain, v[2], v[1]))
					logInfo.Printf("Create %s.%s ------> %s(%s) Success.", v[0], fpsDomain, v[2], v[1])
				}
			}
			i++
			_, err := Db.Exec(sqlStr, domainList["domain"][0]+"."+fpsDomain, domainList["domainAs"][0]+"."+fpsDomain, domainList["domainUs"][0]+"."+fpsDomain, fpsAsNode.id, fpsUsNode.id)
			if err != nil {
				logError.Printf("Create Domain Success, But Sync DB Failed, %s", err)
			}
		}
	}
	return str, nil
}

func GetFpsCode(Db *sql.DB, ip string) string {
	sqlStr := "select code from fps where login_ip = ?"
	var code string
	err := Db.QueryRow(sqlStr, ip).Scan(&code)
	if err != nil {
		logError.Printf("Get Fps Code Failed: %s", err)
	}
	return code
}

func SyncDomainRemark() []string {
	exception := []string{}
	re := regexp.MustCompile(`\.`)
	sqlStr1 := "select id,code,login_ip from fps where status = 1"
	sqlStr2 := "select domain from fps_domain where fps_us_id = ? and status in (1, 2)"
	res, _err := Db.Query(sqlStr1)
	if _err != nil {
		logError.Printf("query failed: %s", _err)
		return nil
	}
	defer res.Close()
	for res.Next() {
		var id int
		var code string
		var ip string
		err := res.Scan(&id, &code, &ip)
		if err != nil {
			logError.Printf("query fps_domain failed: %v", err)
			return nil
		}
		rows, err := Db.Query(sqlStr2, id)
		if err != nil {
			logError.Printf("query failed: %s", err)
			return nil
		}
		defer rows.Close()
		for rows.Next() {
			var domainName string
			err := rows.Scan(&domainName)
			if err != nil {
				logError.Printf("query fps_domain failed: %v", err)
				return nil
			}
			parts := re.Split(domainName, 2)
			res, err := Aliyun.QueryDomainRecord(parts[0], parts[1])
			if err != nil {
				logError.Printf("%s, %s", domainName, err)
			} else {
				if res[0].Remark != code {
					err := Aliyun.UpdateDomainRemark(res[0].RecordId, code)
					if err != nil {
						logError.Printf("Update %s Remark Failed.", domainName)
					} else {
						logInfo.Printf("Update Remark: %s ------> %s", domainName, code)
					}
				}
			}
			if res[0].Value != ip {
				fps_code := GetFpsCode(Db, res[0].Value)
				str := fmt.Sprintf("%s ------> Ali: %s(%s) DB: %s(%s)", domainName, res[0].Value, fps_code, ip, code)
				exception = append(exception, str)
			}
		}
	}
	return exception
}
