package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Response struct {
	Message string      `json:"message" dc:"api tip"`
	Data    interface{} `json:"data"    dc:"api result"`
}

type OcrReq struct {
	g.Meta  `path:"/ocr" method:"post"`
	Content string `v:"required" json:"content"`
}
type OcrRes struct {
	Content string `json:"content" dc:"ocr result"`
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

type Ocr struct{}

func (Ocr) Handler(ctx context.Context, req *OcrReq) (res *OcrRes, err error) {

	decodedBytes, err := base64ToBytes(req.Content)
	if err != nil {
		return nil, err
	}

	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return nil, err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return nil, err
	}
	secret, err := adapter.Get(ctx, "ocr.secret")
	if err != nil {
		return nil, err
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(secret.(string)))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// The Gemini 1.5 models are versatile and work with most use cases
	model := client.GenerativeModel("gemini-1.5-flash")

	prompt := []genai.Part{
		genai.ImageData("png", decodedBytes),
		genai.Text("Get the Content from the picture"),
	}
	resp, err := model.GenerateContent(ctx, prompt...)

	if err != nil {
		return nil, err
	}

	respBytes, _ := json.Marshal(resp.Candidates[0].Content.Parts)
	codes := extractNumbers(string(respBytes))
	res = &OcrRes{}
	if len(codes) > 0 {
		res.Content = codes[0]
	}
	return
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
	r.Response.WriteJson(Response{
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
	s.SetPort(8888)
	s.Run()
}
