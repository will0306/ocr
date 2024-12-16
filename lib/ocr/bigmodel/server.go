package bigmodel

import (
	"codeocr/api"
	"codeocr/lib/tool"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

func extractNumbers(input string) []string {
	re := regexp.MustCompile(`\d+`)

	numbers := re.FindAllString(input, -1)

	return numbers
}

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

type BigModelServ struct{}

func (b BigModelServ) ImageNumber(ctx context.Context, imageBase64 string, modelName string) (resp string, err error) {

	if modelName == "" {
		modelName = "glm-4v-flash"
	}
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return "", err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return "", err
	}
	secret, err := adapter.Get(ctx, "bigmodel.secret")
	if err != nil {
		return "", err
	}

	url := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	method := "POST"

	reqBody := `
{
    "model": "%s",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "image_url",
            "image_url": {
                "url": "%s"
            }
          },
          {
            "type": "text",
            "text": "只返回数字"
          }
        ]
      }
    ]
}
	`
	reqBody = fmt.Sprintf(reqBody, modelName, imageBase64)

	payload := strings.NewReader(reqBody)
	client := &http.Client{}
	httpReq, err := http.NewRequest(method, url, payload)

	if err != nil {
		g.Log().Errorf(ctx, "http_error: %s", err.Error())
		return
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))
	httpReq.Header.Add("Content-Type", "application/json")

	startTime := time.Now().Unix()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		g.Log().Errorf(ctx, "http_request: %s", err.Error())
		return
	}
	endTime := time.Now().Unix()
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		g.Log().Errorf(ctx, "io_ReadAll: %s", err.Error())
		return
	}
	var bigModelResp *api.BigModelResp
	err = json.Unmarshal(body, &bigModelResp)
	if err != nil {
		return "", err
	}
	if len(bigModelResp.Choices) == 0 || bigModelResp.Choices[0].Message.Content == "" {
		g.Log().Warningf(ctx, "%s resp: %+v", tool.GetFuncInfo(), bigModelResp)
		return "", nil
	}
	g.Log().Infof(ctx, "%s cost %d second, ocr: %+v", tool.GetFuncInfo(), endTime-startTime, bigModelResp.Choices[0].Message.Content)
	codes := extractNumbers(bigModelResp.Choices[0].Message.Content)
	return codes[0], nil
}

func (b BigModelServ) PassportInfo(ctx context.Context, imageBase64, modelName string) (resp *api.PassportInfo, err error) {

	if modelName == "" {
		modelName = "glm-4v-flash"
	}
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return nil, err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return nil, err
	}
	secret, err := adapter.Get(ctx, "bigmodel.secret")
	if err != nil {
		return nil, err
	}

	url := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	method := "POST"

	reqBody := `
{
    "model": "%s",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "image_url",
            "image_url": {
                "url": "%s"
            }
          },
          {
            "type": "text",
            "text": "用英文json格式返回出生日期(birth_date)、姓(surname, 字母大写)、名(givename, 字母大写)、护照号(passport_no)、发行日(issue_date)、过期日(expiry_date)、性别(sex, 只有F或者M)、国籍(nationality)、国家代号(country_code), 日期格式: 23/01/1994, 不需要patronymic name"
          }
        ]
      }
    ]
}
	`
	reqBody = fmt.Sprintf(reqBody, modelName, imageBase64)

	payload := strings.NewReader(reqBody)
	client := &http.Client{}
	httpReq, err := http.NewRequest(method, url, payload)

	if err != nil {
		g.Log().Errorf(ctx, "http_error: %s", err.Error())
		return
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))
	httpReq.Header.Add("Content-Type", "application/json")

	startTime := time.Now().Unix()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		g.Log().Errorf(ctx, "http_request: %s", err.Error())
		return
	}
	endTime := time.Now().Unix()
	g.Log().Infof(ctx, "%s cost %d second", tool.GetFuncInfo(), endTime-startTime)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		g.Log().Errorf(ctx, "io_ReadAll: %s", err.Error())
		return
	}
	var bigModelResp *api.BigModelResp
	err = json.Unmarshal(body, &bigModelResp)
	if err != nil {
		return nil, err
	}
	if len(bigModelResp.Choices) == 0 || bigModelResp.Choices[0].Message.Content == "" {
		return nil, nil
	}
	// 使用正则表达式提取JSON内容
	re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
	matches := re.FindStringSubmatch(bigModelResp.Choices[0].Message.Content)
	if len(matches) <= 1 {
		g.Log().Warningf(ctx, "exception, input: %s", bigModelResp.Choices[0].Message.Content)
		return nil, nil
	}
	var passportInfo *api.PassportInfo
	err = json.Unmarshal([]byte(matches[1]), &passportInfo)

	if err != nil {
		return nil, err
	}

	outputFormat := "02/01/2006"
	if converDate, err := parseAndFormatDate(passportInfo.IssueDate, outputFormat); err == nil {
		passportInfo.IssueDate = converDate
	}
	if converDate, err := parseAndFormatDate(passportInfo.ExpiryDate, outputFormat); err == nil {
		passportInfo.ExpiryDate = converDate
	}
	if converDate, err := parseAndFormatDate(passportInfo.BirthDate, outputFormat); err == nil {
		passportInfo.BirthDate = converDate
	}
	return passportInfo, nil
}
