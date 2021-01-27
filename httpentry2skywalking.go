package cls_skywalking_client_go

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"bytes"
	"io/ioutil"

	"cls_skywalking_client_go/util"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"net/http"
	"github.com/labstack/echo/v4"
)

var GRPCReporter go2sky.Reporter
var GRPCTracer *go2sky.Tracer

const componentIDGOHttpServer = 5005

func UseSkyWalking(e *echo.Echo, serviceName string) go2sky.Reporter {
	useSkywalking :=os.Getenv("USE_SKYWALKING")
	if(useSkywalking !="true") {
		return nil
	}

	newReporter, err := getReporter(os.Getenv("USE_SKYWALKING_DEBUG"))
	if err != nil {
		log.Printf("new reporter error %v \n", err)
	} else {
		GRPCReporter = newReporter

		reporter := GRPCReporter
		if reporter == nil {
			return GRPCReporter
		}

		tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(reporter))
		if err != nil {
			log.Printf("create tracer error %v \n", err)
		}

		if tracer != nil {
			GRPCTracer = tracer
		}
	}

	e.Use(LogToSkyWalking)
	go ClearContextAtRegularTime()
	return GRPCReporter
}

func getReporter(isDebug string) (go2sky.Reporter, error) {
	if isDebug == "true" {
		return reporter.NewGRPCReporter("127.0.0.1:8050")
	} else {
		return reporter.NewGRPCReporter("skywalking-oap:11800")
	}
}

func StartLogForCron(e *echo.Echo, taskName string) go2sky.Span {
	if(GRPCTracer == nil) {
		return nil
	}
	c := e.NewContext(nil, nil)
	c.Set("tracer", GRPCTracer)
	safeHeader := make(http.Header)
	safeHeader.Set(propagation.Header, "")
	c.Set("header", newSafeHeader(safeHeader))
	SetContext(c)

	request, err := http.NewRequest("GET", fmt.Sprintf("do_task_%s", taskName), strings.NewReader("暂无参数"))
	if err != nil {
		return nil
	}

	c.SetRequest(request)

	span, ctx, err := GRPCTracer.CreateEntrySpan(c.Request().Context(),
		fmt.Sprintf("do_task_%s", taskName),
		func() (string, error) {
			return c.Get("header").(*SafeHeader).Get(propagation.Header), nil
		})
	if err != nil {
		return nil
	}

	span.SetComponent(componentIDGOHttpServer)
	span.Tag(go2sky.TagHTTPMethod, "GET")
	span.Tag(go2sky.TagURL, taskName)
	span.SetSpanLayer(v3.SpanLayer_Http)
	c.SetRequest(c.Request().WithContext(ctx))

	//span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("请求来源:%s", "test",))
	Log("[开始定时任务]" +  fmt.Sprintf("任务名称:%s,", taskName))

	return span
}

func EndLogForCron(span go2sky.Span,  taskName, result string) {
	if GRPCTracer == nil || span == nil {
		return
	}
	Log("[结束定时任务]" +  fmt.Sprintf("任务名称:%s, 结果:", taskName, result))
	span.End()
}

func LogToSkyWalking(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if GRPCTracer == nil {
			log.Printf("tracer is nil")
			err = next(c)
			return
		}
		c.Set("tracer", GRPCTracer)
		c.Set("header", newSafeHeader(c.Request().Header))
		SetContext(c)
		defer DeleteContext()
		//c.Set("advo", c.Request().Body.AdVo)
		requestUrlArray := strings.Split(c.Request().RequestURI, "?")
		requestParams := getRequestParams(requestUrlArray)

		var requestParamMap = make(map[string]string) /*创建集合 */
		if len(requestParams) != 0 {
			requestParamArray := strings.Split(requestParams, "&")
			for requestParamIndex := range requestParamArray {
				requestParamKeyValue := strings.Split(requestParamArray[requestParamIndex], "=")
				if len(requestParamKeyValue) >= 2 {
					requestParamMap[requestParamKeyValue[0]] = requestParamKeyValue[1]
				}
			}
		}

		span, ctx, err := GRPCTracer.CreateEntrySpan(c.Request().Context(),
			getoperationName(c, requestParamMap, requestUrlArray),
			func() (string, error) {
				return c.Get("header").(*SafeHeader).Get(propagation.Header), nil
			})

		requestParamMap = nil

		if err != nil {
			err = next(c)
			return
		}

		span.SetComponent(componentIDGOHttpServer)
		span.Tag(go2sky.TagHTTPMethod, c.Request().Method)
		span.Tag(go2sky.TagURL, c.Request().Host+c.Request().URL.Path)
		span.SetSpanLayer(v3.SpanLayer_Http)
		c.SetRequest(c.Request().WithContext(ctx))

		bodyBytes, _ := ioutil.ReadAll(c.Request().Body)
		c.Request().Body.Close() //  这里调用Close
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("请求来源:%s,请求参数:%+v, \r\n payload  : %s", c.Request().RemoteAddr,
			requestParams, string(bodyBytes)))
		//	span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("开始请求,请求地址:%s,",  c.Request().RequestURI))

		defer func() {
			code := c.Response().Status
			if code >= 400 {
				span.Error(time.Now(), "Error on handling request")
			}
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
			span.End()
		}()

		err = next(c)
		return
	}
}

func getoperationName(c echo.Context, requestParamMap map[string]string, requestUrlArray []string) string {
	if requestParamMap["os"] == "" {
		return fmt.Sprintf("%s%s", c.Request().Method, util.ReplaceNumber(requestUrlArray[0]))
	} else {
		return fmt.Sprintf("/%s__%s%s", requestParamMap["os"], c.Request().Method, util.ReplaceNumber(requestUrlArray[0]))
	}
}

func getRequestParams(requestUrlArray []string) string {
	condition := len(requestUrlArray) > 1
	if condition {
		return requestUrlArray[1]
	}
	return ""
}
