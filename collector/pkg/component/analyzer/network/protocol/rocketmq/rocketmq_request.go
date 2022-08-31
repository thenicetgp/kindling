package rocketmq

import (
	"log"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRocketMQRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 8
	}
}

func parseRocketMQRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		// 解析报文内容
		var (
			payloadLength int32
			code          int32  //请求码
			language      string //请求方的语言类型、如JAVA。
			//version     int32 //给通信层知道对方的版本号，响应方可以以此做兼容老版本等的特殊操作。
			opaque int32  //标识码,标识为哪一次请求，响应方会原值返回此标识码
			flag   int32  //标识位 if request flag==0; else if response flag==1
			remark string //存放错误信息的文本信息,方便开发人员定位
			//extFieldsMap  map[string]string //自定义字段，不同请求会有不同参数，也可以为空
		)

		message.ReadInt32(0, &payloadLength)
		if payloadLength <= 8 {
			return false, true
		}
		message.ReadInt32(8, &code)
		log.Printf("rocketmq_code is %d",code)
		message.ReadString(12, false, &language)
		log.Printf("rocketmq_language is %s",language)
		if !IsValidLanguage(language) {
			return false, true
		}
		message.ReadInt32(20, &opaque)
		log.Printf("rocketmq_opaque is %d",opaque)
		message.ReadInt32(24, &flag)
		if flag != 0 { //flag=0 request，flag=1 response
			return false, true
		}
		message.ReadString(28, false, &remark)

		// 通过AddStringAttribute() 或 AttIntAttribute() 存储解析出的属性
		message.AddIntAttribute(constlabels.RocketMQRequestCode, int64(code))
		message.AddIntAttribute(constlabels.RocketMQOpaque, int64(opaque))
		message.AddStringAttribute(constlabels.RocketMQRequestRemark, remark)
		// 解析成功
		return true, false
	}
}
