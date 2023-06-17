package GetValue

import (
	"github.com/valyala/fasthttp"
	"strconv"
)

func GetBool(ctx *fasthttp.RequestCtx, valueName string) (bool, error) {
	str := string(ctx.QueryArgs().Peek(valueName))
	var val bool
	var err error

	if str == "" {
		return false, nil
	}
	val, err = strconv.ParseBool(str)
	if err != nil {
		return false, err
	}

	return val, nil
}

func GetInt(ctx *fasthttp.RequestCtx, valueName string) (int, error) {
	ValueStr := string(ctx.QueryArgs().Peek(valueName))
	var value int
	var err error

	if ValueStr == "" {
		return 0, nil
	}

	value, err = strconv.Atoi(ValueStr)
	if err != nil {
		return -1, err
	}

	return value, nil
}
