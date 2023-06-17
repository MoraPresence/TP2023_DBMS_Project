package json

import "github.com/valyala/fasthttp"

func AppJson(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Content-Type", "application/json")
		next(ctx)
	}
}
