package method

import (
	"JumpServerRobot/auth"
	"Tencloud"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"util"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type MessageText struct {
	Text string `json:"text"`
}

type Controller struct {
	MsgID        string
	User         string
	Action       string
	ActionStr    string
	Remark       string
	ResponseText string
	ErrMsg       error
}

var (
	EventHandler *dispatcher.EventDispatcher
	logFile      = "/data/kdl/log/opsServer/jumpServerRobot.log"
	logInfo      = util.LogConf(logFile, "[INFO] ")
	logError     = util.LogConf(logFile, "[ERROR] ")
	client       *lark.Client
)

func init() {
	EventHandler = dispatcher.NewEventDispatcher(auth.Tok, auth.EncryptKey)
	client = lark.NewClient(auth.AppID, auth.AppSecret)
}

func buildData(messID, action, actionStr, remark string) *Controller {
	return &Controller{
		messID,
		"",
		action,
		actionStr,
		remark,
		"",
		nil,
	}
}

func (c *Controller) addIP() {
	c.ResponseText, c.ErrMsg = Tencloud.AddWhiteIP(c.ActionStr, c.Remark)
}

func (c *Controller) handleProcess() {
	switch c.Action {
	case "添加IP", "addIP":
		c.addIP()
	default:
		c.ErrMsg = errors.New("无效指令")
	}
}

func (c *Controller) responseMessage() {
	var respText = struct {
		Text string `json:"text"`
	}{""}
	if c.ErrMsg != nil {
		respText.Text = c.ErrMsg.Error()
	} else {
		respText.Text = c.ResponseText
	}
	t, _ := json.Marshal(respText)
	req := larkim.NewReplyMessageReqBuilder().MessageId(c.MsgID).Body(larkim.NewReplyMessageReqBodyBuilder().Content(string(t)).MsgType(`text`).ReplyInThread(true).Build()).Build()
	resp, err := client.Im.Message.Reply(context.Background(), req)
	if err != nil {
		logError.Println(err)
		return
	}

	if !resp.Success() {
		logError.Println(resp.Code, resp.Msg, resp.RequestId())
		return
	}
}

func ReceiveMessage() {
	EventHandler.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		go func(mesID, userID string) {
			var mess MessageText
			err1 := json.Unmarshal([]byte(*event.Event.Message.Content), &mess)
			if err1 != nil {
				logError.Println(err1)
				return
			}
			newController, err := build(mesID, userID, mess.Text)
			if err != nil {
				return
			}
			if newController.ErrMsg != nil {
				newController.responseMessage()
				return
			}
			newController.handleProcess()
			if newController.ErrMsg != nil {
				newController.responseMessage()
				return
			} else {
				newController.responseMessage()
				logInfo.Printf("%s | %s %s", newController.User, newController.Action, newController.ActionStr)
			}
		}(*event.Event.Message.MessageId, *event.Event.Sender.SenderId.UserId)
		return nil
	})
}

func safeGet(slice []string, index int) string {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return ""
}

func build(messId, userId, req string) (*Controller, error) {
	// str := strings.Join(strings.Fields(req), " ")
	data := strings.Split(strings.Join(strings.Fields(req), " "), " ")
	ms := safeGet(data, 0)
	if ms == "@_all" {
		return nil, errors.New("接收到错误信息")
	}
	action := safeGet(data, 1)
	actionStr := safeGet(data, 2)
	remark := safeGet(data, 3)
	newController := buildData(messId, action, actionStr, remark)
	userAUth := auth.GetUserName(userId)
	if userAUth != "" {
		newController.User = userAUth
	} else {
		newController.ErrMsg = errors.New("当前用户没有权限操作此机器人")
	}
	return newController, nil
}
