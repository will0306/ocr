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
	Platform string `json:"platform" d:"bigmodel"`
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

type OcrDrivingLicenseReq struct {
	g.Meta   `path:"/ocr/driving-license" method:"post"`
	Content  string `json:"content"`
	Url      string `json:"url"`
	Platform string `json:"platform"`
	Model    string `json:"model"`
	Language string `json:"language" d:"en"`
}

type OcrDrivingLicenseRes struct {
	DrivingLicenseInfo *DriverLicenseInfo `json:"driving_license_info"    dc:"api result"`
}

// Point 定义一个二维坐标点
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// KeyValueInfo 定义了 "prism_keyValueInfo" 列表中每个元素的结构
// 它包含了某个识别出的字段、它的值以及在图片中的位置坐标
type KeyValueInfo struct {
	Key   string  `json:"key"`   // 字段的键名 (例如 "name")
	Value string  `json:"value"` // 识别出的字段值 (例如 "张三")
	Pos   []Point `json:"pos"`   // 字段在图片中的位置坐标（通常是一个四边形）
}

// DrivingLicenseFace 对应驾驶证正面（face）识别结果的结构体
type DrivingLicenseFace struct {
	LicenseNumber    string `json:"licenseNumber"`    // 证号
	Name             string `json:"name"`             // 姓名
	Sex              string `json:"sex"`              // 性别
	Nationality      string `json:"nationality"`      // 国籍
	Address          string `json:"address"`          // 住址
	BirthDate        string `json:"birthDate"`        // 出生日期
	InitialIssueDate string `json:"initialIssueDate"` // 初次领证日期
	ApprovedType     string `json:"approvedType"`     // 准驾类型
	IssueAuthority   string `json:"issueAuthority"`   // 发证单位
	ValidFromDate    string `json:"validFromDate"`    // 有效起始日期
	ValidPeriod      string `json:"validPeriod"`      // 有效期限
}

// DrivingLicenseBack 对应驾驶证反面（back）识别结果的结构体
type DrivingLicenseBack struct {
	Name          string `json:"name"`          // 姓名
	RecordNumber  string `json:"recordNumber"`  // 档案编号
	Record        string `json:"record"`        // 记录
	LicenseNumber string `json:"licenseNumber"` // 证号
}

// DrivingLicenseData 对应 "data" 字段，包含正反面的结构化信息
type DrivingLicenseData struct {
	Face DrivingLicenseFace `json:"face"` // 正面信息
	Back DrivingLicenseBack `json:"back"` // 反面信息
}

// DrivingLicenseAPIResponse 是整个驾驶证识别 API 的顶层返回结构
type DrivingLicenseAPIResponse struct {
	Data              *DrivingLicenseData `json:"data"`               // 结构化信息，正面为 face 字段，反面为 back 字段。
	SliceRect         []Point             `json:"sliceRect"`          // 检测出的卡片/子图坐标信息。
	PrismKeyValueInfo []KeyValueInfo      `json:"prism_keyValueInfo"` // 各个字段的坐标信息。
	FType             int                 `json:"ftype"`              // 是否为复印件 (1: 是, 0: 否)。
	Angle             int                 `json:"angle"`              // 图片的角度，0 表示正向, 90 表示图片朝右, 180 朝下, 270 朝左。
	Height            int                 `json:"height"`             // 算法矫正图片后的高度。
	Width             int                 `json:"width"`              // 算法矫正图片后的宽度。
	OrgHeight         int                 `json:"orgHeight"`          // 原图的高度。
	OrgWidth          int                 `json:"orgWidth"`           // 原图的宽度。
}

type DriverLicenseInfo struct {
	Name          string `json:"name"`
	LicenseNumber string `json:"license_number"`
	DateOfBirth   string `json:"date_of_birth"`
	IssueDate     string `json:"issue_date"`
	ExpiryDate    string `json:"expiry_date"`
	Address       string `json:"address"`
	Class         string `json:"class"`
	Gender        string `json:"gender"`
}
