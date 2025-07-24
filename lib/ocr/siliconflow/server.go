package siliconflow

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

var (
	defaultModel = "Qwen/Qwen2-VL-7B-Instruct"
	secretKey    = "siliconflow.secret"
	endPoint     = "https://api.siliconflow.cn/v1/chat/completions"
)

type SiliconflowServ struct{}

func (b SiliconflowServ) ImageNumber(ctx context.Context, imageBase64 string, modelName string) (resp string, err error) {

	if modelName == "" {
		modelName = defaultModel
	}
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return "", err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return "", err
	}
	secret, err := adapter.Get(ctx, secretKey)
	if err != nil {
		return "", err
	}

	url := endPoint
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
            "text": "Return only the number from the image"
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
	codes := tool.ExtractNumbers(bigModelResp.Choices[0].Message.Content)
	if len(codes) == 0 {
		return "", nil
	}
	return codes[0], nil
}

func (b SiliconflowServ) PassportInfo(ctx context.Context, imageBase64, modelName string) (resp *api.PassportInfo, err error) {

	if modelName == "" {
		modelName = defaultModel
	}
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return nil, err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return nil, err
	}
	secret, err := adapter.Get(ctx, secretKey)
	if err != nil {
		return nil, err
	}

	url := endPoint
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
            "text": "Return in English JSON format: birth_date, surname (uppercase letters), givename (uppercase letters), passport_no, issue_date, expiry_date, sex (only F or M), nationality, country_code. Date format: 23/01/1994. Do not include patronymic name."
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
	re := regexp.MustCompile(`(?s)\{.*?\}`)
	matches := re.FindStringSubmatch(bigModelResp.Choices[0].Message.Content)
	if len(matches) == 0 {
		g.Log().Warningf(ctx, "exception, input: %s", bigModelResp.Choices[0].Message.Content)
		return nil, nil
	}
	var passportInfo *api.PassportInfo
	err = json.Unmarshal([]byte(matches[0]), &passportInfo)

	if err != nil {
		return nil, err
	}

	outputFormat := "02/01/2006"
	if converDate, err := tool.ParseAndFormatDate(passportInfo.IssueDate, outputFormat); err == nil {
		passportInfo.IssueDate = converDate
	}
	if converDate, err := tool.ParseAndFormatDate(passportInfo.ExpiryDate, outputFormat); err == nil {
		passportInfo.ExpiryDate = converDate
	}
	if converDate, err := tool.ParseAndFormatDate(passportInfo.BirthDate, outputFormat); err == nil {
		passportInfo.BirthDate = converDate
	}
	return passportInfo, nil
}

func (b SiliconflowServ) DrivingLicenseInfo(ctx context.Context, imageBase64, modelName string) (resp *api.DrivingLicenseAPIResponse, err error) {

	if modelName == "" {
		modelName = defaultModel
	}
	adapter, err := gcfg.NewAdapterFile("config")
	if err != nil {
		return nil, err
	}
	err = adapter.AddPath("config/")
	if err != nil {
		return nil, err
	}
	secret, err := adapter.Get(ctx, secretKey)
	if err != nil {
		return nil, err
	}

	url := endPoint
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
            "text": "Return in English JSON format: {\"data\": {\"face\": {\"licenseNumber\": \"\", \"name\": \"\", \"sex\": \"\", \"nationality\": \"\", \"address\": \"\", \"birthDate\": \"\", \"initialIssueDate\": \"\", \"approvedType\": \"\", \"issueAuthority\": \"\", \"validFromDate\": \"\", \"validPeriod\": \"\"}, \"back\": {\"name\": \"\", \"recordNumber\": \"\", \"record\": \"\", \"licenseNumber\": \"\"}}}. Use uppercase where appropriate. Date format: 02/01/2006. If information not present, leave empty string."
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
	re := regexp.MustCompile(`(?s)\\{.*?\\}`)
	matches := re.FindStringSubmatch(bigModelResp.Choices[0].Message.Content)
	var drivingInfo *api.DrivingLicenseAPIResponse
	if len(matches) == 0 {
		g.Log().Warningf(ctx, "exception, input: %s", bigModelResp.Choices[0].Message.Content)
		err = json.Unmarshal([]byte(bigModelResp.Choices[0].Message.Content), &drivingInfo)
	} else {
		err = json.Unmarshal([]byte(matches[0]), &drivingInfo)
	}

	if err != nil {
		return nil, err
	}

	outputFormat := "02/01/2006"
	if drivingInfo.Data != nil {
		if converDate, err := tool.ParseAndFormatDate(drivingInfo.Data.Face.BirthDate, outputFormat); err == nil {
			drivingInfo.Data.Face.BirthDate = converDate
		}
		if converDate, err := tool.ParseAndFormatDate(drivingInfo.Data.Face.InitialIssueDate, outputFormat); err == nil {
			drivingInfo.Data.Face.InitialIssueDate = converDate
		}
		if converDate, err := tool.ParseAndFormatDate(drivingInfo.Data.Face.ValidFromDate, outputFormat); err == nil {
			drivingInfo.Data.Face.ValidFromDate = converDate
		}
	}
	return drivingInfo, nil
}
