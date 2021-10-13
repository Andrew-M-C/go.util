package conf

// E 表示错误类型
type E string

// Error 实现 error 接口
func (e E) Error() string {
	return string(e)
}

const (
	// ErrParseOptions 标志配置不合法
	ErrParseOptions = E("parse options error")
	// ErrReadFile 表示读取文件失败
	ErrReadFile = E("read file error")
	// ErrFileNotExist 表示配置文件不存在
	ErrFileNotExist = E("file not exist")
	// ErrTarget 表示解析配置的目标错误
	ErrTarget = E("illegal target")
	// ErrWriteToFile 写文件错误
	ErrWriteToFile = E("write to file error")
	// ErrUnmarshal 反序列化错误
	ErrUnmarshal = E("unmarshal error")
)
