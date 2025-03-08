package auth

var AuthUser = map[string]string{
	"": "",
}

var (
	Tok        = ""
	EncryptKey = ""
	AppID      = ""
	AppSecret  = ""
)

func CheckUser(userID string) (r bool) {
	for id := range AuthUser {
		if id == userID {
			r = true
		}
	}
	return
}
