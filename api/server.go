package api

import "github.com/gogf/gf/v2/frame/g"

type BigModelResp struct {
	Choices   []BigModelChoices `json:"choices,omitempty"`
	Created   int               `json:"created,omitempty"`
	ID        string            `json:"id,omitempty"`
	Model     string            `json:"model,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Usage     BigModelUsage     `json:"usage,omitempty"`
}
type BigModelMessage struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}
type BigModelChoices struct {
	FinishReason string          `json:"finish_reason,omitempty"`
	Index        int             `json:"index,omitempty"`
	Message      BigModelMessage `json:"message,omitempty"`
}
type BigModelUsage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type Response struct {
	Message string      `json:"message" dc:"api tip"`
	Data    interface{} `json:"data"    dc:"api result"`
}

type OcrReq struct {
	g.Meta   `path:"/ocr" method:"post"`
	Content  string `v:"required" json:"content"`
	Platform string `json:"platform"`
	Model    string `json:"model"`
}

type OcrRes struct {
	Content string `json:"content" dc:"ocr result"`
}

type OcrPassportReq struct {
	g.Meta   `path:"/ocr/passport" method:"post"`
	Content  string `json:"content"`
	Url      string `json:"url"`
	Platform string `json:"platform"`
	Model    string `json:"model"`
}

type OcrPassportRes struct {
	PassportInfo *PassportInfo `json:"passport_info"    dc:"api result"`
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
