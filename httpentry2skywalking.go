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
	"net/http"
	"reflect"

	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/cls-skywalking-client-go.git/util"
	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/propagation"
	"codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/reporter"
	v3 "codehub-cn-east-2.devcloud.huaweicloud.com/jgz00001/go2sky.git/reporter/grpc/language-agent"

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

	request, err := http.NewRequest("GET", fmt.Sprintf("do_task_%s", taskName), strings.NewReader("暂无参数"))
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

	//span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("请求来源:%s", "test",))
	Log("[开始定时任务]" + fmt.Sprintf("任务名称:%s,", taskName))

	return span
}

func EndLogForCron(span go2sky.Span, taskName, result string) {
	if GRPCTracer == nil || span == nil {
		return
	}
	Log("[结束定时任务]" + fmt.Sprintf("任务名称:%s, 结果:%s", taskName, result))
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
		c.Request().Body.Close() //  这里调用Close
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("请求来源:%s,请求参数:%+v, \r\n payload  : %s", c.Request().RemoteAddr,
			requestParams, string(bodyBytes)))
		//	span.Log(time.Now(), "[HttpRequest]", fmt.Sprintf("开始请求,请求地址:%s,",  c.Request().RequestURI))

		logWithSearchUseRequestParamMap(requestParamMap)

		requestParamMap = nil

		err = next(c)

		defer func() {
			code := c.Response().Status
			if code >= 400 {
				span.Error(time.Now(), fmt.Sprintf("code:%s,  Error on handling request", strconv.Itoa(code)))
			}
			if err != nil {
				errorStr := fmt.Sprintf("code:%s, 错误响应： %#v", strconv.Itoa(code), err)
				needFilter := filter(errorStr)
				if needFilter {
					span.Log(time.Now(), errorStr)
				} else {
					span.Error(time.Now(), errorStr)
				}
			}

			if c.Response().Size <= 1000 {
				//200 响应中notFountCode := "Code:404"
				//errno 不为空
				//if()
				logResponse(span, c.Response())
			}

			span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
			span.End()
		}()
		return
	}
}

func filter(str string) bool {
	var list = []string{"code:\"20101\"", "code:\"10212\"", "无审核权限",
		"验证码错误",
		"请登录",
		"该文章正在被审核",
		"非草稿箱内容或者不存在",
		"视频还未处于可正常播放状态",
		"板块不能为空",
		"正在编辑中"}
	for ingorestrIndex := range list {

		if strings.Contains(str, list[ingorestrIndex]) {
			return true
		}
	}
	return false
}

func logResponse(span go2sky.Span, resp *echo.Response) {
	w := resp.Writer

	var bytes []byte
	//支持GZIP
	t := reflect.ValueOf(reflect.ValueOf(w).Elem().FieldByName("compressor"))
	if !isBlank(t) {
		bytes = reflect.ValueOf(w).Elem().FieldByName("w").Elem().FieldByName("w").Elem().FieldByName("buf").Bytes()
	} else {
		bytes = reflect.ValueOf(w).Elem().FieldByName("w").Elem().FieldByName("buf").Bytes()
	}

	str2 := string(bytes[:])
	fmt.Println(str2)

	//data.Errno = 501
	span.Log(time.Now(), str2)
}

func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}


func logWithSearchUseRequestParamMap(requestParamMap map[string]string) string {
	searchableKeys := ""
	if len(requestParamMap["sv"]) != 0 {
		searchableKeys += fmt.Sprintf("sv=%s", requestParamMap["sv"])
	}

	if len(requestParamMap["app"]) != 0 {
		if len(searchableKeys) != 0 {
			searchableKeys += ","
		}
		searchableKeys += fmt.Sprintf("app=%s", requestParamMap["app"])
	}

	if len(requestParamMap["cuid"]) != 0 {
		if len(searchableKeys) != 0 {
			searchableKeys += ","
		}
		searchableKeys += fmt.Sprintf("cuid=%s", requestParamMap["cuid"])
	}
	if len(searchableKeys) != 0 {
		LogWithSearch(searchableKeys, "Input search")
	}

	return searchableKeys
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
