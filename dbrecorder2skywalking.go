package cls_skywalking_client_go

import (
	"github.com/Masterminds/structable"
)

type RecordProxy struct {
	Recorder structable.Recorder
}

// “构造基类”
func NewRecorderProxy(recorder structable.Recorder) *RecordProxy {
	return &RecordProxy{
		Recorder: recorder,
	}
}

func (f RecordProxy) getRecorder() structable.Recorder {
	return f.Recorder
}

func (f *RecordProxy) Insert() error {
	return f.getRecorder().Insert()
}

func (f *RecordProxy) Delete() error {
	return f.getRecorder().Delete()
}

func (f *RecordProxy) Update() error {
	return f.getRecorder().Update()
}

func (f *RecordProxy) Load() error {
	return f.getRecorder().Load()
}

func (f *RecordProxy) LoadWhere(arg1 interface{}, arg2 ...interface{}) error {
	return f.getRecorder().LoadWhere(arg1, arg2)
}

func (f *RecordProxy) ExistsWhere(arg1 interface{}, arg2 ...interface{}) (bool, error) {
	return f.getRecorder().ExistsWhere(arg1, arg2)
}


//func (f RecordProxy) Insert() (*sql.Rows, error) {
//	queryStr, args, _ := query.ToSql()
//
//	var temp = make([]string, len(args))
//	for k, v := range args {
//		temp[k] = fmt.Sprintf("%d", v)
//	}
//	var result = "[" + strings.Join(temp, ",") + "]"
//
//	reqSpan, spanErr := StartSpantoSkyWalkingForDb(queryStr+"\r\n Parameters: "+result, os.Getenv("DB_URL"))
//	if spanErr != nil {
//		log.Printf("StartSpantoSkyWalkingForDb error: %v \n", spanErr)
//	}
//
//	rows, err := query.RunWith(f.getDb()).Query()
//
//	if err != nil {
//		EndSpantoSkywalkingForDb(reqSpan, queryStr, false, err)
//	}
//
//	EndSpantoSkywalkingForDb(reqSpan, queryStr, true, err)
//
//	return rows, err
//}
//
//func StartSpantoSkyWalkingForDb(queryStr string, db string) (go2sky.Span, error) {
//	originCtx := GetContext()
//	if originCtx == nil {
//		return nil, errors.New("can not get context")
//	}
//	ctx := originCtx.(echo.Context)
//	// op_name 是每一个操作的名称
//	tracerFromCtx := ctx.Get("tracer")
//	if tracerFromCtx == nil {
//		return nil, errors.New("can not get tracer")
//	}
//	tracer := tracerFromCtx.(*go2sky.Tracer)
//	reqSpan, err := tracer.CreateExitSpan(ctx.Request().Context(), queryStr, db, func(header string) error {
//		ctx.Get("header").(*SafeHeader).Set(propagation.Header, header)
//		return nil
//	})
//	reqSpan.SetComponent(5)
//	reqSpan.SetSpanLayer(v3.SpanLayer_Database) // rpc 调用
//	reqSpan.Log(time.Now(), "[DBRecord Request]", fmt.Sprintf("开始请求,请求服务:%s,请求地址:%s", db, queryStr))
//
//	return reqSpan, err
//}
//
//func EndSpantoSkywalkingForDb(reqSpan go2sky.Span, queryStr string, isNormal bool, err error) {
//	if reqSpan == nil {
//		return
//	}
//	reqSpan.Tag(go2sky.TagDBType, "MySql")
//	reqSpan.Tag(go2sky.TagURL, queryStr)
//	if !isNormal {
//		reqSpan.Error(time.Now(), "[DBRecord Response]", fmt.Sprintf("结束请求,响应结果: %s", err))
//	} else {
//		reqSpan.Log(time.Now(), "[DBRecord Response]", "结束请求")
//	}
//	reqSpan.End()
//}
