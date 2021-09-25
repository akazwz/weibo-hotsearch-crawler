package notify

import (
	"encoding/json"
	"fmt"
	"github.com/akazwz/weibo-hotsearch-crawler/global"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func SendMessage() {
	credential := common.NewCredential(global.CFG.SecretId, global.CFG.SecretKey)

	cpf := profile.NewClientProfile()
	cpf.SignMethod = "HmacSHA1"

	client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)

	request := sms.NewSendSmsRequest()

	request.SmsSdkAppId = common.StringPtr("1400576425")
	request.SignName = common.StringPtr("赵文卓工作学习")
	request.SenderId = common.StringPtr("")
	request.ExtendCode = common.StringPtr("")
	request.TemplateParamSet = common.StringPtrs([]string{})
	request.TemplateId = common.StringPtr("1131592")
	request.PhoneNumberSet = common.StringPtrs([]string{"+8615153953308"})

	response, err := client.SendSms(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
	if err != nil {
		panic(err)
	}
	b, _ := json.Marshal(response.Response)
	// 打印返回的json字符串
	fmt.Printf("%s", b)
}
