package main

import (
	"codeocr/api"
	"codeocr/lib/ocr"
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
)

type Ocr struct{}

func (Ocr) OcrHandler(ctx context.Context, req *api.OcrReq) (res *api.OcrRes, err error) {

	serv := ocr.NewOcr(req.Platform)
	resp, err := serv.ImageNumber(ctx, req.Content, req.Model)
	if err != nil {
		return nil, err
	}
	res = &api.OcrRes{
		Content: resp,
	}
	return res, nil
}

func (Ocr) PassportHandler(ctx context.Context, req *api.OcrPassportReq) (resp *api.OcrPassportRes, err error) {

	serv := ocr.NewOcr(req.Platform)
	passportInfo, err := serv.PassportInfo(ctx, req.Content, req.Model)
	if err != nil {
		return nil, err
	}
	resp = &api.OcrPassportRes{
		PassportInfo: passportInfo,
	}
	return resp, nil

}

func Middleware(r *ghttp.Request) {
	r.Middleware.Next()

	var (
		msg string
		res = r.GetHandlerResponse()
		err = r.GetError()
	)
	if err != nil {
		msg = err.Error()
	} else {
		msg = "OK"
	}
	r.Response.WriteJson(api.Response{
		Message: msg,
		Data:    res,
	})
}

func main() {
	s := g.Server()
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(Middleware)
		group.Bind(
			new(Ocr),
		)
	})

	ctx := gctx.New()
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return
	}
	err = adapter.AddPath("config/")
	if err != nil {
		g.Log().Error(ctx, err.Error())
		return
	}
	s.Run()
}
