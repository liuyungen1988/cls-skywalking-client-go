package cls_skywalking_client_go

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/labstack/echo"
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
	newReporter, err := reporter.NewGRPCReporter("127.0.0.1:8050")

	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
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
			return
		}

		tracer, err := go2sky.NewTracer(TransHttpEntry.ServiceName, go2sky.WithReporter(reporter))
		if err != nil {
			log.Fatalf("create tracer error %v \n", err)
		}

		if tracer == nil {
			err = next(c)
			return
		}

		c.Set("tracer", tracer)

		span, ctx, err := tracer.CreateEntrySpan(c.Request().Context(),
			fmt.Sprintf("/%s%s", c.Request().Method, strings.Split(c.Request().RequestURI, "?")[0]),
			func() (string, error) {
				return c.Request().Header.Get(propagation.Header), nil
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
