package ocr

import (
	"codeocr/api"
	"codeocr/lib/ocr/bigmodel"
	"codeocr/lib/ocr/gemini"
	"codeocr/lib/ocr/mistral"
	"codeocr/lib/ocr/openrouter"
	"codeocr/lib/ocr/siliconflow"
	"context"
)

var (
	defaultPlatform = "gemini"
	platformMap     = map[string]OcrServer{
		"gemini":      gemini.GeminiServ{},
		"bigmodel":    bigmodel.BigModelServ{},
		"mistral":     mistral.MistralServ{},
		"openrouter":  openrouter.OpenRouterServ{},
		"siliconflow": siliconflow.SiliconflowServ{},
	}
)

type OcrServer interface {
	ImageNumber(ctx context.Context, imageBase64, modelName string) (resp string, err error)
	PassportInfo(ctx context.Context, imageBase64, modelName string) (resp *api.PassportInfo, err error)
}

func NewOcr(platform string) (serv OcrServer) {
	if serv, ok := platformMap[platform]; !ok {
		return platformMap[defaultPlatform]
	} else {
		return serv
	}
}
