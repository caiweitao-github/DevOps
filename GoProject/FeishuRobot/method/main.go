package method

import (
	"FeishuRobot/auth"
	fh "FeishuRobot/handle"
	"Tencloud"
	tyYun "TianYiYun"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	UserID       string
	Action       string
	ActionStr    string
	Remark       string
	ResponseText string
	ErrMsg       error
}

var (
	EventHandler *dispatcher.EventDispatcher
	logFile      = "/data/kdl/log/opsServer/feishuRobot.log"
	logInfo      = util.LogConf(logFile, "[INFO] ")
	logError     = util.LogConf(logFile, "[ERROR] ")
	client       *lark.Client
)

func init() {
	EventHandler = dispatcher.NewEventDispatcher(auth.Tok, auth.EncryptKey)
	client = lark.NewClient(auth.AppID, auth.AppSecret)
}

func buildData(messID, userId, action, actionStr, remark string) *Controller {
	return &Controller{
		messID,
		userId,
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

func (c *Controller) deleteIP() {
	c.ResponseText, c.ErrMsg = Tencloud.DeleteWhiteIP(c.ActionStr)
}

func (c *Controller) getIP() {
	c.ResponseText, c.ErrMsg = Tencloud.GetWhiteIP()
}

func (c *Controller) checkKpsNode() {
	c.ResponseText, c.ErrMsg = fh.CheckKpsNode()
}

func (c *Controller) checkSfpsNode() {
	c.ResponseText, c.ErrMsg = fh.CheckSfpsNode()
}

func (c *Controller) kpsRenewHandle() {
	c.ResponseText, c.ErrMsg = tyYun.RenewHandle(c.ActionStr)
}

func (c *Controller) tpsDomainChange() {
	c.ResponseText, c.ErrMsg = fh.UpdateTpsDomainRecord(c.ActionStr, c.Remark)
}

// func (c *Controller) hunyunAI() {
// 	c.ResponseText, c.ErrMsg = Tencloud.HunYuan(c.ActionStr)
// }

// func (c *Controller) testProcess() {
// 	time.Sleep(15 * time.Second)
// 	c.ResponseText = "success."
// 	c.ErrMsg = nil
// }

func (c *Controller) helpDoc() {
	text := fmt.Sprintf("尊敬的用户%s, ", auth.AuthUser[c.UserID])
	c.ResponseText = text + `机器人目前支持功能如下:
添加JumpServer白名单IP ➡️ 添加IP|addIP (e.g. addIP 1.1.1.1)

删除JumpServerIP ➡️ 删除IP|deleteIP (e.g. deleteIP 1.1.1.1)

查看JumpServer白名单 ➡️ 查看IP|getIP

检查节点是否到期 ➡️ 检查sfps节点|检查kps节点

kps节点续费 ➡️ kps续费 (e.g. kps续费 kps888)

切换TPS域名 ➡️ 切换TPS域名 (e.g. 切换TPS域名 tid tpsxx)`
}

func (c *Controller) handleProcess() {
	switch c.Action {
	case "添加IP", "addIP":
		c.addIP()
	case "删除IP", "deleteIP":
		c.deleteIP()
	case "查看IP", "getIP":
		c.getIP()
	case "检查kps节点":
		c.checkKpsNode()
	case "检查sfps节点":
		c.checkSfpsNode()
	case "kps续费":
		c.kpsRenewHandle()
	case "切换TPS域名":
		c.tpsDomainChange()
	// case "提问":
	// 	c.hunyunAI()
	case "帮助":
		c.helpDoc()
	// case "test":
	// 	c.testProcess()
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
				logInfo.Printf("%s | %s %s", auth.AuthUser[newController.UserID], newController.Action, newController.ActionStr)
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
	data := strings.Split(strings.Join(strings.Fields(req), " "), " ")
	ms := safeGet(data, 0)
	if ms == "@_all" {
		return nil, errors.New("接收到错误信息")
	}
	action := safeGet(data, 1)
	actionStr := safeGet(data, 2)
	remark := safeGet(data, 3)
	newController := buildData(messId, userId, action, actionStr, remark)
	userAUth := auth.CheckUser(userId)
	if !userAUth {
		newController.ErrMsg = errors.New("当前用户没有权限操作此机器人")
	}
	return newController, nil
}
