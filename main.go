package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
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

type OcrPassportReq struct {
	g.Meta  `path:"/ocr/passport" method:"post"`
	Content string `v:"required" json:"content"`
	Url     string `json:"url"`
}

type OcrPassportRes struct {
	Message string        `json:"message" dc:"api tip"`
	Data    *PassportInfo `json:"data"    dc:"api result"`
}

type PassportInfo struct {
	BirthDate   string `json:"birth_date"`
	Surname     string `json:"surname"`
	Givename    string `json:"givename"`
	PassportNo  string `json:"passport_no"`
	IssueDate   string `json:"issue_date"`
	ExpiryDate  string `json:"expiry_date"`
	Sex         string `json:"sex"`
	Nationality string `json:"nationality"`
	CountryCode string `json:"country_code"`
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

func (Ocr) OcrHandler(ctx context.Context, req *OcrReq) (res *OcrRes, err error) {

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

func (Ocr) PassportHandler(ctx context.Context, req *OcrPassportReq) (resp *OcrPassportRes, err error) {

	// 转换 Base64 字符串为 io.Reader
	reader, err := base64ToReader(req.Content)
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

	startTime := time.Now().Unix()
	// The Gemini 1.5 models are versatile and work with most use cases
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFile(ctx, "", reader, nil)
	if err != nil {
		return nil, err
	}
	defer client.DeleteFile(ctx, file.Name)

	geminiResp, err := model.GenerateContent(ctx, genai.Text("用英文json格式返回出生日期(birth_date)、姓(surname, 字母大写)、名(givename, 字母大写)、护照号(passport_no)、发行日(issue_date)、过期日(expiry_date)、性别(sex, 只有F或者M)、国籍(nationality)、国家代号(country_code), 日期格式: 23/01/1994, 不需要patronymic name "), genai.FileData{URI: file.URI})
	endTime := time.Now().Unix()
	g.Log().Infof(ctx, "PassportHandler cost %d second", endTime-startTime)
	if err != nil {
		return nil, err
	}

	// 定义用于匹配 JSON 的正则表达式
	re := regexp.MustCompile(`\{.*?\}`)

	// 查找所有 JSON 子串
	partsBytes, _ := json.Marshal(geminiResp.Candidates[0].Content.Parts)

	jsonStrings := re.FindAllString(string(partsBytes), -1)
	// 输出提取的 JSON 字符串
	var passportInfo *PassportInfo
	for _, jsonString := range jsonStrings {
		jsonString = strings.ReplaceAll(jsonString, "\\n", "")
		jsonString = strings.ReplaceAll(jsonString, "\\", "")

		err := json.Unmarshal([]byte(jsonString), &passportInfo)
		if err != nil {
			return nil, err
		}
		break
	}
	resp = &OcrPassportRes{
		Data: passportInfo,
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
