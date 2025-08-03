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

func (b SiliconflowServ) DrivingLicenseInfo(ctx context.Context, imageBase64, modelName, language string) (resp *api.DriverLicenseInfo, err error) {
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
            "text": "Extract the following fields from this driver's license image: Name, license_number, date_of_birth (format yyyy.mm.dd), issue_date (format yyyy.mm.dd), expiry_date (format yyyy.mm.dd), Address, Class, Gender. Return as JSON object."
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
		return nil, err
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))
	httpReq.Header.Add("Content-Type", "application/json")

	startTime := time.Now().Unix()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		g.Log().Errorf(ctx, "http_request: %s", err.Error())
		return nil, err
	}
	endTime := time.Now().Unix()
	g.Log().Infof(ctx, "%s cost %d second", tool.GetFuncInfo(), endTime-startTime)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		g.Log().Errorf(ctx, "io_ReadAll: %s", err.Error())
		return nil, err
	}
	var bigModelResp *api.BigModelResp
	err = json.Unmarshal(body, &bigModelResp)
	if err != nil {
		return nil, err
	}
	if len(bigModelResp.Choices) == 0 || bigModelResp.Choices[0].Message.Content == "" {
		return nil, nil
	}
	content := bigModelResp.Choices[0].Message.Content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	var jsonStr string
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonStr = content[startIdx : endIdx+1]
	} else {
		jsonStr = content
	}
	var info api.DriverLicenseInfo
	err = json.Unmarshal([]byte(jsonStr), &info)
	if err != nil {
		return nil, err
	}
	if language != "" && language != "English" {
		transReqBody := `
{
    "model": "%s",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "Translate to %s ONLY the following fields: Name '%s' to translated Name, Address '%s' to translated Address, Class '%s' to translated Class, Gender '%s' to translated Gender. For the other fields, COPY EXACTLY: License Number '%s', Date of Birth '%s', Issue Date '%s', Expiry Date '%s'. Return a JSON object with fields: name, license_number, date_of_birth, issue_date, expiry_date, address, class, gender using the translated or copied values."
          }
        ]
      }
    ]
}
		`
		transReqBody = fmt.Sprintf(transReqBody, modelName, language, info.Name, info.Address, info.Class, info.Gender, info.LicenseNumber, info.DateOfBirth, info.IssueDate, info.ExpiryDate)
		transPayload := strings.NewReader(transReqBody)
		transHttpReq, err := http.NewRequest(method, url, transPayload)
		if err != nil {
			return &info, nil
		}
		transHttpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))
		transHttpReq.Header.Add("Content-Type", "application/json")
		transHttpResp, err := client.Do(transHttpReq)
		if err != nil {
			return &info, nil
		}
		defer transHttpResp.Body.Close()
		transBody, err := io.ReadAll(transHttpResp.Body)
		if err != nil {
			return &info, nil
		}
		var transBigModelResp *api.BigModelResp
		err = json.Unmarshal(transBody, &transBigModelResp)
		if err != nil {
			return &info, nil
		}
		if len(transBigModelResp.Choices) == 0 || transBigModelResp.Choices[0].Message.Content == "" {
			return &info, nil
		}
		transContent := transBigModelResp.Choices[0].Message.Content
		transStartIdx := strings.Index(transContent, "{")
		transEndIdx := strings.LastIndex(transContent, "}")
		if transStartIdx != -1 && transEndIdx != -1 && transEndIdx > transStartIdx {
			transJsonStr := transContent[transStartIdx : transEndIdx+1]
			err = json.Unmarshal([]byte(transJsonStr), &info)
			if err != nil {
				return &info, nil
			}
		}
	}
	outputFormat := "2006.01.02"
	if converted, err := tool.ParseAndFormatDate(info.DateOfBirth, outputFormat); err == nil {
		info.DateOfBirth = converted
	}
	if converted, err := tool.ParseAndFormatDate(info.IssueDate, outputFormat); err == nil {
		info.IssueDate = converted
	}
	if converted, err := tool.ParseAndFormatDate(info.ExpiryDate, outputFormat); err == nil {
		info.ExpiryDate = converted
	}
	return &info, nil
}
