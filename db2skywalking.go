package cls_skywalking_client_go

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
	"log"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/labstack/echo"
	"errors"
)

type DbProxy struct {
	Db sq.DBProxyBeginner
}

// “构造基类”
func NewDbProxy(db sq.DBProxyBeginner) *DbProxy {
	return &DbProxy{
		Db: db,
	}
}

func (f DbProxy) getDb() sq.DBProxyBeginner {
	return f.Db
}

func (f DbProxy) Query(ctx echo.Context, query squirrel.SelectBuilder) (*sql.Rows, error) {
	queryStr, args, _ := query.ToSql()

	var temp = make([]string, len(args))
	for k, v := range args {
		temp[k] = fmt.Sprintf("%d", v)
	}
	var result = "[" + strings.Join(temp, ",") + "]"

	reqSpan, spanErr := StartSpantoSkyWalkingForDb(ctx, queryStr+ "\r\n Parameters: " + result, os.Getenv("DB_URL"))
	if spanErr != nil {
		log.Fatalf("StartSpantoSkyWalkingForDb error: %v \n", spanErr)
	}

	rows, err := query.RunWith(f.getDb()).Query()

	if err != nil {
		EndSpantoSkywalkingForDb(reqSpan, queryStr, false, err)
	}

	EndSpantoSkywalkingForDb(reqSpan, queryStr, true, err)

	return rows, err
}

func StartSpantoSkyWalkingForDb(ctx echo.Context, queryStr string, db string) (go2sky.Span, error) {
	// op_name 是每一个操作的名称
	tracerFromCtx := ctx.Get("tracer")
	if tracerFromCtx == nil  {
		return nil,  errors.New("can not get tracer")
	}
	tracer := tracerFromCtx.(*go2sky.Tracer)
	reqSpan, err := tracer.CreateExitSpan(ctx.Request().Context(), queryStr, db, func(header string) error {
		ctx.Request().Header.Set(propagation.Header, header)
		return nil
	})
	reqSpan.SetComponent(5)
	reqSpan.SetSpanLayer(v3.SpanLayer_Database) // rpc 调用
	reqSpan.Log(time.Now(), "[DBRequest]", fmt.Sprintf("开始请求,请求服务:%s,请求地址:%s", db, queryStr))

	return reqSpan, err
}

func EndSpantoSkywalkingForDb(reqSpan go2sky.Span, queryStr string, isNormal bool, err error) {
	if reqSpan == nil {
		return
	}
	reqSpan.Tag(go2sky.TagDBType, "MySql")
	reqSpan.Tag(go2sky.TagURL, queryStr)
	if !isNormal {
		reqSpan.Error(time.Now(), "[DB Response]", fmt.Sprintf("结束请求,响应结果: %s", err))
	} else {
		reqSpan.Log(time.Now(), "[DB Response]", "结束请求")
	}
	reqSpan.End()
}
