package cls_skywalking_client_go

import (
	"net/http"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/labstack/echo"
	"fmt"
)

func StartSpantoSkyWalking(ctx echo.Context, req *http.Request, url string, params []string, remoteService string) (go2sky.Span, error) {
	// op_name 是每一个操作的名称
	tracer := ctx.Get("tracer").(*go2sky.Tracer)
	reqSpan, err := tracer.CreateExitSpan(ctx.Request().Context(), url, remoteService, func(header string) error {
		req.Header.Set(propagation.Header, header)
		return nil
	})
	reqSpan.SetComponent(2)                 //HttpClient,看 https://github.com/apache/skywalking/blob/master/docs/en/guides/Component-library-settings.md ， 目录在component-libraries.yml文件配置
	reqSpan.SetSpanLayer(v3.SpanLayer_Http) // rpc 调用
	reqSpan.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("开始请求,请求服务:%s,请求地址:%s,请求参数:%+v", remoteService, url, params))

	return reqSpan, err
}

func EndSpantoSkywalking(reqSpan go2sky.Span, url string, resp string, isNormal bool, err error) {
	reqSpan.Tag(go2sky.TagHTTPMethod, http.MethodPost)
	reqSpan.Tag(go2sky.TagURL, url)
	if !isNormal {
		reqSpan.Error(time.Now(), "[HttpRequest]", fmt.Sprintf("结束请求,返回异常: %s", err.Error()))
	} else {
		reqSpan.Log(time.Now(), "[Http Response]", fmt.Sprintf("结束请求,响应结果: %s", resp))
	}
	reqSpan.End()
}
