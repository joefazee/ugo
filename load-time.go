package ugo

import (
	"regexp"
	"runtime"
	"time"
)

func (u *Ugo) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	pc, _, _, _ := runtime.Caller(1)

	funcObj := runtime.FuncForPC(pc)
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	u.InfoLog.Printf("Load Time: %s took %s\n", name, elapsed)
}
