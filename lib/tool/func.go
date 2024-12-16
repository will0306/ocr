package tool

import (
	"runtime"
	"strings"
)

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
