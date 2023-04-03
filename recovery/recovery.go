// Package recovery 提供 recover() 函数及相关功能的封装
package recovery

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Andrew-M-C/go.util/log"
	"github.com/Andrew-M-C/go.util/runtime/caller"
)

// CatchPanic 捕获异常。需要用 defer 函数调用
func CatchPanic(opts ...Option) {
	e := recover()
	if e == nil {
		return
	}

	o := mergeOptions(opts)
	stack := caller.GetAllCallers()
	stack = stack[2:]
	for len(stack) > 0 {
		// 查找 stack 直至找到业务代码
		s := stack[0]
		if strings.HasSuffix(string(s.File), "/runtime/panic.go") {
			stack = stack[1:]
		} else {
			break
		}
	}

	if o.withErrorLog {
		panicDesc, _ := json.Marshal(fmt.Sprint(e))
		stackDesc, _ := json.Marshal(stack)

		bdr := strings.Builder{}
		bdr.WriteString("caught panic, information: '")
		bdr.Write(panicDesc)
		bdr.WriteString("', stack information: '")
		bdr.Write(stackDesc)
		bdr.WriteByte('\'')

		if o.ctx != nil {
			log.ErrorContext(o.ctx, bdr.String())
		} else {
			log.Error(bdr.String())
		}
	}

	if o.callback != nil {
		o.callback(stack)
	}
}
