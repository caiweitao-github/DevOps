package main

import (
	"JumpServerRobot/auth"
	r "JumpServerRobot/method"
	"context"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func main() {
	cli := larkws.NewClient(auth.AppID, auth.AppSecret,
		larkws.WithEventHandler(r.EventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)
	r.ReceiveMessage()
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
