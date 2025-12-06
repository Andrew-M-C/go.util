package simpledb

// E 错误类型
type E string

func (e E) Error() string {
	return string(e)
}

const (
	// ErrColumnDuplicate 表示列值已存在了（设置了唯一约束的时候）
	ErrColumnDuplicate = E("column duplicate")
	// ErrEmptyLineKey 给定的行名为 ""
	ErrEmptyLineKey = E("line key is empty")
)
