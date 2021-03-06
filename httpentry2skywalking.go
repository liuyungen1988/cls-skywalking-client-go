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

	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/cls-skywalking-client-go.git/util"
	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/propagation"
	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/reporter"
	v3 "codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/reporter/grpc/language-agent"
	"net/http"

	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git"
	"github.com/labstack/echo/v4"
)

var GRPCReporter go2sky.Reporter
var GRPCTracer *go2sky.Tracer

const componentIDGOHttpServer = 5005

func UseSkyWalking(e *echo.Echo, serviceName string) go2sky.Reporter {
	useSkywalking := os.Getenv("USE_SKYWALKING")
	if useSkywalking != "true" {
		return nil
	}

	newReporter, err := getReporter(os.Getenv("USE_SKYWALKING_DEBUG"), os.Getenv("SKYWALKING_OAP_IP"))
	if err != nil {
		log.Printf("new reporter error %v \n", err)
	} else {
		GRPCReporter = newReporter

		reporter := GRPCReporter
		if reporter == nil {
			return GRPCReporter
		}

		sample := 1.0
		sampleStr := os.Getenv("USE_SKYWALKING_SAMPLE")
		if len(sampleStr) != 0 {
			sample, _ = strconv.ParseFloat(sampleStr, 64)
		}

		tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(reporter), go2sky.WithSampler(sample))
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

func getReporter(isDebug string, skywalkingOapIp string) (go2sky.Reporter, error) {
	if len(skywalkingOapIp) != 0 {
		return reporter.NewGRPCReporter(skywalkingOapIp)
	}

	if isDebug == "true" {
		return reporter.NewGRPCReporter("127.0.0.1:8050")
	} else {
		return reporter.NewGRPCReporter("skywalking-oap:11800")
	}
}

func StartLogForCron(e *echo.Echo, taskName string) go2sky.Span {
	if GRPCTracer == nil {
		return nil
	}
	c := e.NewContext(nil, nil)
	c.Set("tracer", GRPCTracer)
	safeHeader := make(http.Header)
	safeHeader.Set(propagation.Header, "")
	c.Set("header", newSafeHeader(safeHeader))
	SetContext(c)

	request, err := http.NewRequest("GET", fmt.Sprintf("do_task_%s", taskName), strings.NewReader("????????????"))
	if err != nil {
		return nil
	}

	c.SetRequest(request)

	span, ctx, err := GRPCTracer.CreateEntrySpan(c.Request().Context(),
		fmt.Sprintf("do_task_%s", taskName),
		func() (string, error) {
			value := ""
			if c.Get("header") != nil {
				value = c.Get("header").(*SafeHeader).Get(propagation.Header)
			}
			return value, nil
		})
	if err != nil {
		return nil
	}

	span.SetComponent(componentIDGOHttpServer)
	span.Tag(go2sky.TagHTTPMethod, "GET")
	span.Tag(go2sky.TagURL, taskName)
	span.SetSpanLayer(v3.SpanLayer_Http)
	c.SetRequest(c.Request().WithContext(ctx))

	//span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("????????????:%s", "test",))
	Log("[??????????????????]" + fmt.Sprintf("????????????:%s,", taskName))

	return span
}

func EndLogForCron(span go2sky.Span, taskName, result string) {
	if GRPCTracer == nil || span == nil {
		return
	}
	Log("[??????????????????]" + fmt.Sprintf("????????????:%s, ??????:", taskName, result))
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

		var requestParamMap = make(map[string]string) /*???????????? */
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
				value := ""
				if c.Get("header") != nil {
					value = c.Get("header").(*SafeHeader).Get(propagation.Header)
				}
				return value, nil
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
		c.Request().Body.Close() //  ????????????Close
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("????????????:%s,????????????:%+v, \r\n payload  : %s", c.Request().RemoteAddr,
			requestParams, string(bodyBytes)))
		//	span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("????????????,????????????:%s,",  c.Request().RequestURI))

		if len(requestParamMap["sv"]) != 0 {
			searchableKeys := fmt.Sprintf("sv=%s", requestParamMap["sv"])
			LogWithSearch(searchableKeys, "Input sv search")
		}

		if len(requestParamMap["app"]) != 0 {
			searchableKeys := fmt.Sprintf("app=%s", requestParamMap["app"])
			LogWithSearch(searchableKeys, "Input app search")
		}

		requestParamMap = nil

		err = next(c)

		defer func() {
			code := c.Response().Status
			if code >= 400 {
				span.Error(time.Now(), fmt.Sprintf("code:%s,  Error on handling request", strconv.Itoa(code)))
			}
			if err != nil {
				span.Error(time.Now(), fmt.Sprintf("code:%s, ??????????????? %#v", strconv.Itoa(code), err))
			}

			span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
			span.End()
		}()
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
