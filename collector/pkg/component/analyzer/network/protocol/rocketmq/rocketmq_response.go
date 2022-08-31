package rocketmq

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRocketMQResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		// 根据报文实现具体的fastFail()函数
		return len(message.Data) < 8
	}
}

func parseRocketMQResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		//解析响应报文
		var (
			payloadLength int32
			code          int32  //响应码
			language      string //请求方的语言类型、如JAVA。
			flag          int32  //flag=0 request，flag=1 response
			opaque        int32  //原请求的opaque，应答方不做修改，原值返回
			remark        string //应答的文本信息，通常存放错误信息
			//extFieldsMap  map[string]string //自定义字段，不同响应会有不同参数，也可以为空
		)

		message.ReadInt32(0, &payloadLength)
		if payloadLength <= 8 {
			return false, true
		}
		message.ReadInt32(8, &code)
		message.ReadString(12, false, &language)
		if !IsValidLanguage(language) {
			return false, true
		}

		if len(message.Data) < 20 {
			return false, true
		}
		message.ReadInt32(20, &opaque)
		if !message.HasAttribute(constlabels.RocketMQOpaque) ||
			message.GetIntAttribute(constlabels.RocketMQOpaque) != int64(opaque) {

			return false, true
		}

		message.ReadInt32(24, &flag)
		if flag != 1 { //flag=0 request，flag=1 response
			return false, true
		}
		message.ReadString(28, false, &remark)

		//message.Offset = 32
		message.AddIntAttribute(constlabels.RocketMQResponseCode, int64(code))
		message.AddStringAttribute(constlabels.RocketMQResponseRemark, remark)
		// 解析成功
		return true, true
	}
}
