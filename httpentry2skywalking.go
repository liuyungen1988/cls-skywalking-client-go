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

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/labstack/echo/v4"
)

var GRPCReporter go2sky.Reporter
var TransHttpEntry *HttpEntry

type HttpEntry struct {
	ServiceName string
}

// “构造基类”
func NewHttpEntry(serviceName string) *HttpEntry {
	return &HttpEntry{
		ServiceName: serviceName,
	}
}

const componentIDGOHttpServer = 5005

func UseSkyWalking(e *echo.Echo, serviceName string) go2sky.Reporter {
	useSkywalking :=os.Getenv("USE_SKYWALKING")
	if(useSkywalking !="true") {
		return nil
	}
	newReporter, err := reporter.NewGRPCReporter("skywalking-oap:11800")
	if err != nil {
		log.Printf("new reporter error %v \n", err)
	} else {
		GRPCReporter = newReporter
	}

	initHttpEntry := NewHttpEntry(serviceName)
	if initHttpEntry != nil {
		TransHttpEntry = initHttpEntry
	}

	e.Use(LogToSkyWalking)
	return GRPCReporter
}

func LogToSkyWalking(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		reporter := GRPCReporter
		if reporter == nil {
			err = next(c)
			return
		}

		tracer, err := go2sky.NewTracer(TransHttpEntry.ServiceName, go2sky.WithReporter(reporter))
		if err != nil {
			log.Printf("create tracer error %v \n", err)
		}

		if tracer == nil {
			err = next(c)
			return
		}

		c.Set("tracer", tracer)
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

		span, ctx, err := tracer.CreateEntrySpan(c.Request().Context(),
			getoperationName(c, requestParamMap, requestUrlArray),
			func() (string, error) {
				return c.Get("header").(*SafeHeader).Get(propagation.Header), nil
			})

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
		return fmt.Sprintf("%s%s", c.Request().Method, requestUrlArray[0])
	} else {
		return fmt.Sprintf("/%s__%s%s", requestParamMap["os"], c.Request().Method, requestUrlArray[0])
	}
}

func getRequestParams(requestUrlArray []string) string {
	condition := len(requestUrlArray) > 1
	if condition {
		return requestUrlArray[1]
	}
	return ""
}
