package tool

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func ExtractNumbers(input string) []string {
	re := regexp.MustCompile(`\d+`)

	numbers := re.FindAllString(input, -1)

	return numbers
}

// parseAndFormatDate 尝试将多种日期格式转换为 dd/mm/yyyy 格式
func ParseAndFormatDate(inputDate string, outputFormat string) (string, error) {
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

func GetFuncInfo() string {
	skip := 2
	size := 6
	pc := make([]uintptr, size)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])

	frame, _ := frames.Next()
	s := strings.Split(frame.Function, ".")
	return strings.Join(s, ":")
}
