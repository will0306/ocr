package gemini

import (
	"bytes"
	"codeocr/api"
	"codeocr/lib/tool"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// isHTTPLink 判断字符串是否是 HTTP/HTTPS 链接
func isHTTPLink(s string) bool {
	// 正则表达式匹配 HTTP/HTTPS 链接
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(s)
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

type GeminiServ struct{}

func (b GeminiServ) ImageNumber(ctx context.Context, imageBase64 string, modelName string) (resp string, err error) {

	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	decodedBytes, err := base64ToBytes(imageBase64)
	if err != nil {
		return "", err
	}

	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return "", err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return "", err
	}
	secret, err := adapter.Get(ctx, "ocr.secret")
	if err != nil {
		return "", err
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(secret.(string)))
	if err != nil {
		return "", err
	}
	defer client.Close()

	// The Gemini 1.5 models are versatile and work with most use cases
	genaiModel := client.GenerativeModel(modelName)

	prompt := []genai.Part{
		genai.ImageData("png", decodedBytes),
		genai.Text("Get the Number from the picture"),
	}
	startTime := time.Now().Unix()
	genaiResp, err := genaiModel.GenerateContent(ctx, prompt...)
	endTime := time.Now().Unix()

	if err != nil {
		return "", err
	}

	respBytes, _ := json.Marshal(genaiResp.Candidates[0].Content.Parts)
	g.Log().Infof(ctx, "%s cost %d second, ocr: %s", tool.GetFuncInfo(), endTime-startTime, respBytes)
	codes := extractNumbers(string(respBytes))
	return codes[0], nil
}

func (b GeminiServ) PassportInfo(ctx context.Context, imageBase64, modelName string) (resp *api.PassportInfo, err error) {

	var reader io.Reader
	var imageBytes []byte
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	if isHTTPLink(imageBase64) {
		imageResp, err := http.Get(imageBase64)
		if err != nil {
			return nil, err
		}
		defer imageResp.Body.Close()
		reader = imageResp.Body
		imageBytes, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
	} else {

		reader, err = base64ToReader(imageBase64)
		if err != nil {
			return nil, err
		}
		imageBytes, err = io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
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
		return nil, err
	}
	defer client.Close()

	startTime := time.Now().Unix()
	// The Gemini 1.5 models are versatile and work with most use cases
	genaiModel := client.GenerativeModel(modelName)

	genaiReq := []genai.Part{
		genai.ImageData("jpeg", imageBytes),

		genai.Text("用英文json格式返回出生日期(birth_date)、姓(surname, 字母大写)、名(givename, 字母大写)、护照号(passport_no)、发行日(issue_date)、过期日(expiry_date)、性别(sex, 只有F或者M)、国籍(nationality)、国家代号(country_code), 日期格式: 23/01/1994, 不需要patronymic name "),
	}

	// Generate content.
	geminiResp, err := genaiModel.GenerateContent(ctx, genaiReq...)
	if err != nil {
		return nil, err
	}

	endTime := time.Now().Unix()
	g.Log().Infof(ctx, "%s cost %d second", tool.GetFuncInfo(), endTime-startTime)
	if err != nil {
		return nil, err
	}

	// 定义用于匹配 JSON 的正则表达式
	re := regexp.MustCompile(`\{.*?\}`)

	// 查找所有 JSON 子串
	partsBytes, _ := json.Marshal(geminiResp.Candidates[0].Content.Parts)

	jsonStrings := re.FindAllString(string(partsBytes), -1)
	// 输出提取的 JSON 字符串
	var passportInfo *api.PassportInfo
	for _, jsonString := range jsonStrings {
		jsonString = strings.ReplaceAll(jsonString, "\\n", "")
		jsonString = strings.ReplaceAll(jsonString, "\\", "")

		err := json.Unmarshal([]byte(jsonString), &passportInfo)
		if err != nil {
			return nil, err
		}
		break
	}
	return passportInfo, nil
}
