package main

import (
	"bytes"
	"codeocr/api"
	"codeocr/lib/ocr"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
)

// parseAndFormatDate 尝试将多种日期格式转换为 dd/mm/yyyy 格式
func parseAndFormatDate(inputDate string, outputFormat string) (string, error) {
	// 定义多种可能的输入日期格式
	formats := []string{
		"02 Jan 2006",     // 08 JUN 1996
		"02/01/2006",      // 08/06/1996
		"2006-01-02",      // 1996-06-08
		"02-01-2006",      // 08-06-1996
		"02.01.2006",      // 08.06.1996
		"January 2, 2006", // June 8, 1996
		"2 Jan 2006",      // 8 JUN 1996
	}

	// 遍历格式列表并尝试解析
	var parsedTime time.Time
	var err error
	for _, format := range formats {
		parsedTime, err = time.Parse(format, strings.ToUpper(inputDate))
		if err == nil {
			// 成功解析，退出循环
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("无法解析日期: %s", inputDate)
	}

	// 格式化为目标格式 dd/mm/yyyy
	return parsedTime.Format(outputFormat), nil
}

func extractNumbers(input string) []string {
	re := regexp.MustCompile(`\d+`)

	numbers := re.FindAllString(input, -1)

	return numbers
}

func base64ToBytes(base64Data string) ([]byte, error) {
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("Base64 decode error: %v", err)
	}

	return data, nil
}

func base64ToReader(base64Str string) (io.Reader, error) {

	if idx := strings.Index(base64Str, ","); idx != -1 {
		base64Str = base64Str[idx+1:]
	}
	// 解码 Base64 字符串为字节数组
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(data)
	return reader, nil
}

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
