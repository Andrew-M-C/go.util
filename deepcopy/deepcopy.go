// Package deepcopy 用于分析并输出一个 struct 的深复制代码
package deepcopy

// GenerateDeepCopyCode 生成深复制代码
func GenerateDeepCopyCode(prototype interface{}, options ...Option) (string, error) {
	g := newGenerator(prototype)
	for _, o := range options {
		if o != nil {
			o(g)
		}
	}

	if err := g.do(); err != nil {
		return "", err
	}
	return g.packageResult(), nil
}
