package resp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
)

func SendUnexpError(errorMessage string, ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(http.StatusInternalServerError)
	fmt.Printf("{level: error, message: %s}", errors.New(errorMessage))
}

func Send(code int, data interface{}, ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(code)

	serializedData, err := json.Marshal(data)
	if err != nil {
		SendUnexpError(err.Error(), ctx)
		return
	}
	ctx.SetBody(serializedData)
}

func SendOk(data interface{}, ctx *fasthttp.RequestCtx) {
	Send(200, data, ctx)
}
