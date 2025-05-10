package http

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// ReadSSEReadSSEJsonDataData 处理 SSE 事件, 仅处理 'data: ' 行, 并使用 json 反序列化。
//
// Options 目前仅支持 debug 日志
func ReadSSEJsonData[T any](
	ctx context.Context, r io.Reader, handler func(event T),
	opts ...RequestOption,
) error {
	if handler == nil {
		return errors.New("sse handler is nil")
	}
	o := mergeOptions(opts, json.Marshal)

	// 使用 bufio.Reader 读取 SSE 事件流
	reader := bufio.NewReader(r)
	var data string

	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				o.debugf("Got '%v'", err)
				break
			}
			o.debugf("Error: '%v'", err)
			return err
		}

		// 去除行尾的 \n 或 \r\n
		line = strings.TrimRight(line, "\n\r")
		o.debugf("Read line: '%s'", line)

		// 空行表示事件结束，处理收集到的数据
		if line == "" && data != "" {
			var event T
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return err
			}

			// 调用处理函数
			handler(event)
			data = ""
			continue
		}

		// 处理 SSE 数据行
		if strings.HasPrefix(line, "data:") {
			// 提取数据部分
			payload := strings.TrimPrefix(line, "data:")
			// 去除可能存在的前导空格
			payload = strings.TrimPrefix(payload, " ")
			data = payload
		}
	}

	return nil
}
