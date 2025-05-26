package http_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Andrew-M-C/go.util/net/http"
)

func TestReadSSEJsonData(t *testing.T) {
	type event struct {
		Event string `json:"event"`
		Data  string `json:"data"`
	}

	cv("标准 openai 模式", t, func() {
		body := "" +
			`data: {"event": "message", "data": "hello1"}` + "\n\n" +
			`data: {"event": "message", "data": "hello2"}` + "\n\n" +
			`[DONE]` + "\n"
		r := strings.NewReader(body)

		var readEvents []event
		handler := func(e event) {
			readEvents = append(readEvents, e)
		}

		err := http.ReadSSEJsonData(context.Background(), r, handler, http.WithDebugger(t.Logf))
		so(err, isNil)
		so(len(readEvents), eq, 2)
		so(readEvents[0].Data, eq, "hello1")
		so(readEvents[1].Data, eq, "hello2")
	})

	cv("deepseek 的 keep-alive 模式", t, func() {
		// https://api-docs.deepseek.com/zh-cn/quick_start/rate_limit
		body := "" +
			`data: {"event": "message", "data": "hello1"}` + "\n\n" +
			`: keep-alive` + "\n\n" +
			`data: {"event": "message", "data": "hello2"}` + "\n\n"
		r := strings.NewReader(body)

		var readEvents []event
		handler := func(e event) {
			readEvents = append(readEvents, e)
		}

		err := http.ReadSSEJsonData(context.Background(), r, handler, http.WithDebugger(t.Logf))
		so(err, isNil)
		so(len(readEvents), eq, 2)
		so(readEvents[0].Data, eq, "hello1")
		so(readEvents[1].Data, eq, "hello2")
	})

	cv("带 id 的模式", t, func() {
		body := "" +
			`id: 1` + "\n" +
			`data: {"event": "message", "data": "hello1"}` + "\n\n" +
			`id: 2` + "\n" +
			`data: {"event": "message", "data": "hello2"}` + "\n\n"
		r := strings.NewReader(body)

		var readEvents []event
		handler := func(e event) {
			readEvents = append(readEvents, e)
		}

		err := http.ReadSSEJsonData(context.Background(), r, handler, http.WithDebugger(t.Logf))
		so(err, isNil)
		so(len(readEvents), eq, 2)
		so(readEvents[0].Data, eq, "hello1")
		so(readEvents[1].Data, eq, "hello2")
	})

	cv("反序列化错误回调", t, func() {
		body := "" +
			`data: {"event": "message", "data": "hello1"}` + "\n\n" +
			`data: {"event": "message", "data": "hello2"}` + "\n\n" +
			`data: DONE` + "\n"
		r := strings.NewReader(body)

		handler := func(event) { /* do nothing */ }

		cv("没有回调", func() {
			err := http.ReadSSEJsonData(context.Background(), r, handler,
				http.WithDebugger(t.Logf),
			)
			so(err, isErr)
		})

		cv("有回调", func() {
			gotUnmarshalError := false
			errHandler := func(err error, data string) {
				t.Logf("err: %v, data: %s", err, data)
				gotUnmarshalError = true
			}
			err := http.ReadSSEJsonData(context.Background(), r, handler,
				http.WithDebugger(t.Logf), http.WithSSEUnmarshalErrorCallback(errHandler),
			)
			so(err, isNil)
			so(gotUnmarshalError, eq, true)
		})
	})
}
