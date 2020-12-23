package cls_skywalking_client_go

import (
	"fmt"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	v3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/labstack/echo"

	"errors"
	"gopkg.in/redis.v5"
	"strconv"
)

type RedisProxy struct {
	RedisCache *redis.Client
}

// “构造基类”
func NewRedisProxy(redisCache *redis.Client) *RedisProxy {
	return &RedisProxy{
		RedisCache: redisCache,
	}
}

func (f RedisProxy) getRedisCache() *redis.Client {
	return f.RedisCache
}

func (f RedisProxy) Get(ctx echo.Context, key string) *redis.StringCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, "Get "+key, f.getRedisCache().String())

	cmd := f.getRedisCache().Get(key)

	_, err := cmd.Result()
	defer processResult(span, "Get "+key,
		err)
	return cmd
}

func (f RedisProxy) GetRange(ctx echo.Context, key string, start, end int64) *redis.StringCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("GetRange %s, start %s, end %s", key, strconv.FormatInt(start, 10), strconv.FormatInt(end, 10)), f.getRedisCache().String())

	cmd := f.getRedisCache().GetRange(key, start, end)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("GetRange %s, start %s, end %s", key, strconv.FormatInt(start, 10), strconv.FormatInt(end, 10)),
		err)
	return cmd
}

func (f RedisProxy) GetSet(ctx echo.Context, key string, value interface{}) *redis.StringCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, "GetSet "+key, f.getRedisCache().String())

	cmd := f.getRedisCache().GetSet(key, value)

	_, err := cmd.Result()
	defer processResult(span, "GetSet "+key,
		err)
	return cmd
}

func (f RedisProxy) MGet(ctx echo.Context, keys ...string) *redis.SliceCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("MGet %v", keys), f.getRedisCache().String())

	cmd := f.getRedisCache().MGet(keys...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("MGet %v", keys),
		err)
	return cmd
}

func (f RedisProxy) MSet(ctx echo.Context, pairs ...interface{}) *redis.StatusCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("MSet %v", pairs), f.getRedisCache().String())

	cmd := f.getRedisCache().MSet(pairs...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("MSet %v", pairs),
		err)
	return cmd
}

func (f RedisProxy) MSetNX(ctx echo.Context, pairs ...interface{}) *redis.BoolCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("MSetNX %v", pairs), f.getRedisCache().String())

	cmd := f.getRedisCache().MSetNX(pairs...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("MSetNX %v", pairs),
		err)
	return cmd
}

// Redis `SET key value [expiration]` command.
//
// Use expiration for `SETEX`-like behavior.
// Zero expiration means the key has no expiration time.
func (f RedisProxy) Set(ctx echo.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("Set %s, value %v", key,
		value), f.getRedisCache().String())

	cmd := f.getRedisCache().Set(key, value, expiration)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("Set %s, value %v", key,
		value),
		err)
	return cmd
}

func (f RedisProxy) SetRange(ctx echo.Context, key string, offset int64, value string) *redis.IntCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("SetRange %s, offset %s,  value %s", key, strconv.FormatInt(offset, 10), value), f.getRedisCache().String())

	cmd := f.getRedisCache().SetRange(key, offset, value)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("SetRange %s, offset %s,  value %s", key, strconv.FormatInt(offset, 10), value),
		err)
	return cmd
}

func (f RedisProxy) HDel(ctx echo.Context, key string, fields ...string) *redis.IntCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HDel %s, fields %v", key, fields), f.getRedisCache().String())

	cmd := f.getRedisCache().HDel(key, fields...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HDel %s, fields %v", key, fields),
		err)
	return cmd
}

func (f RedisProxy) HExists(ctx echo.Context, key, field string) *redis.BoolCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HExists %s, field %s", key, field), f.getRedisCache().String())

	cmd := f.getRedisCache().HExists(key, field)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HExists %s, field %s", key, field),
		err)
	return cmd
}

func (f RedisProxy) HGet(ctx echo.Context, key, field string) *redis.StringCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HGet %s, field %s", key, field), f.getRedisCache().String())

	cmd := f.getRedisCache().HGet(key, field)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HGet %s, field %s", key, field),
		err)
	return cmd
}

func (f RedisProxy) HGetAll(ctx echo.Context, key string) *redis.StringStringMapCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, "HGetAll  "+key, f.getRedisCache().String())

	cmd := f.getRedisCache().HGetAll(key)

	_, err := cmd.Result()
	defer processResult(span, "HGetAll  "+key,
		err)
	return cmd
}

func (f RedisProxy) HMGet(ctx echo.Context, key string, fields ...string) *redis.SliceCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HMGet %s, fields %v ", key, fields), f.getRedisCache().String())

	cmd := f.getRedisCache().HMGet(key, fields...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HMGet %s, fields %v ", key, fields),
		err)
	return cmd
}

func (f RedisProxy) HMSet(ctx echo.Context, key string, fields map[string]string) *redis.StatusCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, "HMSet  "+key, f.getRedisCache().String())

	cmd := f.getRedisCache().HMSet(key, fields)

	_, err := cmd.Result()
	defer processResult(span, "HMGet  "+key,
		err)
	return cmd
}

func (f RedisProxy) HSet(ctx echo.Context, key, field string, value interface{}) *redis.BoolCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HSet %s, fields %s, value %v ", key, field, value), f.getRedisCache().String())

	cmd := f.getRedisCache().HSet(key, field, value)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HSet %s, fields %s, value %v ", key, field, value),
		err)
	return cmd
}

func (f RedisProxy) HSetNX(ctx echo.Context, key, field string, value interface{}) *redis.BoolCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HSetNX %s, fields %s, value %v ", key, field, value), f.getRedisCache().String())

	cmd := f.getRedisCache().HSetNX(key, field, value)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HSetNX %s, fields %s, value %v ", key, field, value),
		err)
	return cmd
}

func (f RedisProxy) LPop(ctx echo.Context, key string) *redis.StringCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, "LPop  "+key, f.getRedisCache().String())

	cmd := f.getRedisCache().LPop(key)

	_, err := cmd.Result()
	defer processResult(span, "LPop  "+key,
		err)
	return cmd
}

func (f RedisProxy) LPush(ctx echo.Context, key string, values ...interface{}) *redis.IntCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("LPush %s, values %v ", key, values), f.getRedisCache().String())

	cmd := f.getRedisCache().LPush(key, values...)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("LPush %s, values %v ", key, values),
		err)
	return cmd
}

func (f RedisProxy) LPushX(ctx echo.Context, key string, value interface{}) *redis.IntCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("LPushX %s, value %v ", key, value), f.getRedisCache().String())

	cmd := f.getRedisCache().LPushX(key, value)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("LPushX %s, value %v ", key, value),
		err)
	return cmd
}

func (f RedisProxy) LRange(ctx echo.Context, key string, start, stop int64) *redis.StringSliceCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("LRange %s, start %s, stop %s ", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)), f.getRedisCache().String())

	cmd := f.getRedisCache().LRange(key, start, stop)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("LRange %s, start %s, stop %s "+key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)),
		err)
	return cmd
}


func (f RedisProxy) ZRange(ctx echo.Context, key string, start, stop int64) *redis.StringSliceCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("ZRange %s, start %s, stop %s ", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)), f.getRedisCache().String())

	cmd := f.getRedisCache().ZRange(key, start, stop)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("ZRange %s, start %s, stop %s ", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)),
		err)
	return cmd
}

func (f RedisProxy) ZRangeWithScores(ctx echo.Context, key string, start, stop int64) *redis.ZSliceCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("ZRangeWithScores %s, start %s, stop %s ", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)), f.getRedisCache().String())

	cmd := f.getRedisCache().ZRangeWithScores(key, start, stop)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("ZRangeWithScores %s, start %s, stop %s ", key, strconv.FormatInt(start, 10), strconv.FormatInt(stop, 10)),
		err)
	return cmd
}

func (f RedisProxy) HIncrBy(ctx echo.Context, key, field string, incr int64) *redis.IntCmd {
	span, _ := StartSpantoSkyWalkingForRedis(ctx, fmt.Sprintf("HIncrBy %s, field %s, incr %s ", key, field, strconv.FormatInt(incr, 10)), f.getRedisCache().String())

	cmd := f.getRedisCache().HIncrBy(key, field, incr)

	_, err := cmd.Result()
	defer processResult(span, fmt.Sprintf("HIncrBy %s, field %s, incr %s ", key, field, strconv.FormatInt(incr, 10)),
		err)
	return cmd
}

func StartSpantoSkyWalkingForRedis(ctx echo.Context, queryStr string, db string) (go2sky.Span, error) {
	// op_name 是每一个操作的名称
	tracerFromCtx := ctx.Get("tracer")
	if tracerFromCtx == nil {
		return nil,  errors.New("can not get tracer")
	}
	tracer := tracerFromCtx.(*go2sky.Tracer)
	reqSpan, err := tracer.CreateExitSpan(ctx.Request().Context(), queryStr, db, func(header string) error {
		ctx.Request().Header.Set(propagation.Header, header)
		return nil
	})
	reqSpan.SetComponent(7)
	reqSpan.SetSpanLayer(v3.SpanLayer_Cache) // cache
	reqSpan.Log(time.Now(), "[Redis Request]", fmt.Sprintf("开始请求,请求服务:%s,请求地址:%s", db, queryStr))

	return reqSpan, err
}

func EndSpantoSkywalkingForRedis(reqSpan go2sky.Span, queryStr string, isNormal bool, err error) {
	if reqSpan == nil {
		return
	}
	reqSpan.Tag(go2sky.TagURL, queryStr)
	if !isNormal {
		reqSpan.Error(time.Now(), "[Redis Response]", fmt.Sprintf("结束请求,响应结果: %s", err))
	} else {
		reqSpan.Log(time.Now(), "[Redis Response]", "结束请求")
	}
	reqSpan.End()
}

func processResult(span go2sky.Span, str string, err error) {
	if err == nil {
		EndSpantoSkywalkingForRedis(span, str, true, nil)
	} else {
		EndSpantoSkywalkingForRedis(span, str, false, err)
	}
}
