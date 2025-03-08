package Tencloud

import (
	"Tencloud/auth"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

func HunYuan(context string) (string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"
	client, _ := hunyuan.NewClient(auth.Cwt, "", cpf)
	request := hunyuan.NewChatCompletionsRequest()
	request.Model = common.StringPtr("hunyuan-pro")
	request.Messages = []*hunyuan.Message{
		{
			Role:    common.StringPtr("user"),
			Content: common.StringPtr(context),
		},
	}

	response, err := client.ChatCompletions(request)
	if err != nil {
		return "", err
	}
	// if response.Response != nil {
	// 	// 非流式响应
	// 	// fmt.Println(response.ToJsonString())
	// 	fmt.Println(*response.Response.Choices[0].Message.Content)
	// } else {
	// 	// 流式响应
	// 	for event := range response.Events {
	// 		fmt.Println(string(event.Data))
	// 	}
	// }
	return *response.Response.Choices[0].Message.Content, nil
}
