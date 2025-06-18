package procfs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Reference: https://blog.csdn.net/qq_32761549/article/details/135262978
type CPUStat struct {
	User      int64
	Nice      int64
	System    int64
	Idle      int64
	IOWait    int64
	IRQ       int64
	SoftIRQ   int64
	Steal     int64
	Guest     int64
	GuestNice int64
}

func (st *CPUStat) Total() int64 {
	return st.User + st.Nice + st.System + st.Idle + st.IOWait + st.IRQ + st.SoftIRQ + st.Steal + st.Guest + st.GuestNice
}

// ReadCPUStat 读取 /proc/stat 文件内容, 获取 CPU 总的使用情况
func ReadCPUStat() (CPUStat, error) {
	stat := CPUStat{}
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return stat, err
	}

	// 使用 bufio.Scanner 逐行读取内容
	scanner := bufio.NewScanner(bytes.NewReader(content))

	// 查找以 "cpu " 开头的行（注意有空格，这是总体 CPU 统计）
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			// 分割字段
			fields := strings.Fields(line)

			// 检查字段数量，至少需要 11 个字段（cpu + 10个数值）
			if len(fields) < 11 {
				return stat, fmt.Errorf("invalid /proc/stat format: expected at least 11 fields, got %d", len(fields))
			}

			// 解析各个数值字段，跳过第一个 "cpu" 标识符
			values := make([]int64, 10)
			for i := 1; i <= 10; i++ {
				val, err := strconv.ParseInt(fields[i], 10, 64)
				if err != nil {
					return stat, fmt.Errorf("failed to parse field %d ('%s'): %w", i, fields[i], err)
				}
				values[i-1] = val
			}

			// 按照 /proc/stat 文档中的顺序填充结构体
			stat.User = values[0]      // user: normal processes executing in user mode
			stat.Nice = values[1]      // nice: niced processes executing in user mode
			stat.System = values[2]    // system: processes executing in kernel mode
			stat.Idle = values[3]      // idle: twiddling thumbs
			stat.IOWait = values[4]    // iowait: waiting for I/O to complete
			stat.IRQ = values[5]       // irq: servicing interrupts
			stat.SoftIRQ = values[6]   // softirq: servicing softirqs
			stat.Steal = values[7]     // steal: involuntary wait
			stat.Guest = values[8]     // guest: running a normal guest
			stat.GuestNice = values[9] // guest_nice: running a niced guest

			break
		}
	}
	if err := scanner.Err(); err != nil {
		return stat, fmt.Errorf("error reading /proc/stat: %w", err)
	}

	return stat, nil
}
