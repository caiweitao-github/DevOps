package auth

import (
	"database/sql"
	"util"

)

var (
	Db         *sql.DB
	Tok        = ""
	EncryptKey = ""
	AppID      = ""
	AppSecret  = ""
)

func init() {
	var err error
	Db, err = util.KdlStaff()
	if err != nil {
		panic(err)
	}
}

func GetUserName(userID string) (userName string) {
	sqlStr := `select nickname from user_profile where feishu_user_id = ?`
	err := Db.QueryRow(sqlStr, userID).Scan(&userName)
	if err != nil {
		return
	}
	return
}
