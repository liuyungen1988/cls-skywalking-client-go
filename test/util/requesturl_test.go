package util

import (
	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/cls-skywalking-client-go.git/util"
	"fmt"
	"testing"
)

func TestReplaceAccessKeyId1(t *testing.T) {
	url := "http://vod.cn-shanghai.aliyuncs.com/?AccessKeyId=fdfdfdfd&Acti"

	result := util.ReplaceAccessKeyId(url)

	expectecDbUrl := "http://vod.cn-shanghai.aliyuncs.com/?AccessKeyId=***&Acti"

	if result != expectecDbUrl {
		t.Error(fmt.Sprintf("error, result is %s", result))
	}
}

func TestReplaceAccessKeyId2(t *testing.T) {
	url := "http://vod.cn-shanghai.aliyuncs.com/?parama1=fdaf&AccessKeyId=fdfdfdfd&Acti"

	result := util.ReplaceAccessKeyId(url)

	expectecDbUrl := "http://vod.cn-shanghai.aliyuncs.com/?parama1=fdaf?AccessKeyId=***&Acti"

	if result != expectecDbUrl {
		t.Error(fmt.Sprintf("error, result is %s", result))
	}
}


func TestReplaceAccessNumber1(t *testing.T) {
	url := "http://vod.cn-shanghai.aliyuncs.com/123/123?AccessKeyId=fdfdfdfd&Acti"

	result := util.ReplaceNumber(url)

	expectecDbUrl := "http://vod.cn-shanghai.aliyuncs.com/_number_/_number_?AccessKeyId=fdfdfdfd&Acti"

	if result != expectecDbUrl {
		t.Error(fmt.Sprintf("error, result is %s", result))
	}
}
