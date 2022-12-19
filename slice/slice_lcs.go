package slice

// LCSMap 表示 LCS 计算结果图
type LCSMap struct {
	M [][]int
	L [][]LCSLink

	maxX  int
	maxY  int
	equal EqualFunc
}

// LCSLink 表示
type LCSLink struct {
	Down       bool
	LowerRight bool
	Right      bool
}

type EqualFunc func(xIndex, yIndex int) bool

// LCS 计算最长公共子序列，并返回结果图
func LCS(lenX, lenY int, equal EqualFunc) (m *LCSMap) {
	m = makeLCSMap(lenX, lenY)
	if lenX == 0 || lenY == 0 || equal == nil {
		return m
	}

	m.equal = equal

	for i := lenX; i >= 0; i-- {
		for j := lenY; j >= 0; j-- {
			m.ste(i, j)
		}
	}

	return m
}

func makeLCSMap(lx, ly int) *LCSMap {
	m := make([][]int, lx+1)
	l := make([][]LCSLink, lx+1)

	for i := range m {
		m[i] = make([]int, ly+1)
		l[i] = make([]LCSLink, ly+1)
	}

	res := &LCSMap{
		M:    m,
		L:    l,
		maxX: lx,
		maxY: ly,
	}
	return res
}

// ste 状态转移方程
func (m *LCSMap) ste(i, j int) int {
	if i == m.maxX || j == m.maxY {
		m.M[i][j] = 0
		return 0
	}

	if m.equal(i, j) {
		m.M[i][j] = m.M[i+1][j+1] + 1
		m.L[i][j].LowerRight = true
		return m.M[i][j]
	}

	if m.M[i+1][j] == m.M[i][j+1] {
		m.M[i][j] = m.M[i+1][j]
		m.L[i][j].Right, m.L[i][j].Down = true, true

	} else if m.M[i+1][j] > m.M[i][j+1] {
		m.M[i][j] = m.M[i+1][j]
		m.L[i][j].Down = true

	} else {
		m.M[i][j] = m.M[i][j+1]
		m.L[i][j].Right = true
	}

	if m.M[i][j] == 0 {
		m.L[i][j].Right, m.L[i][j].Down, m.L[i][j].LowerRight = false, false, false
	}
	return m.M[i][j]
}

// Route 表示一个结果路径
type Route struct {
	XIndexes []int
	YIndexes []int
}

// GetARoute 方便地返回一个相同序列的路径。从 map 中的右下角，优先向上方搜索
func (m *LCSMap) GetRoute() *Route {
	if m.maxX == 0 || m.maxY == 0 {
		return &Route{}
	}
	if m.M[0][0] == 0 {
		return &Route{}
	}
	res := m.trace(0, 0, &Route{})
	return res[0]
}

// MaxSubLen 返回最大子序列计算结果
func (m *LCSMap) MaxSubLen() int {
	if m.maxX == 0 || m.maxY == 0 {
		return 0
	}
	return m.M[0][0]
}

func (m *LCSMap) trace(i, j int, r *Route) (res []*Route) {
	// 等于0了，结束
	if m.M[i][j] == 0 {
		return []*Route{r}
	}

	// 有一个相等的节点，append
	if m.L[i][j].LowerRight {
		r.XIndexes = append(r.XIndexes, i)
		r.YIndexes = append(r.YIndexes, j)
		// logf(f, "%p - process: (%d, %d) r.XIndexes: %v, r.YIndexes: %v", r, i, j, r.XIndexes, r.YIndexes)
		return m.trace(i+1, j+1, r)
	}

	// 没有相等的节点，那么看看有没有分叉
	if m.L[i][j].Down && !m.L[i][j].Right {
		res = m.trace(i+1, j, r)
	} else if !m.L[i][j].Down && m.L[i][j].Right {
		res = m.trace(i, j+1, r)
	} else {
		res = m.trace(i, j+1, r)
		// res = m.trace(i+1, j, r)
	}
	return res
}
