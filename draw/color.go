package draw

// Red 表示简单红色
const (
	White = WhiteColor(0xFF)
	Black = BlackColor(0xFF)
	Red   = RedColor(0xFF)
	Green = GreenColor(0xFF)
	Blue  = BlueColor(0xFF)
)

type RedColor uint8
type GreenColor uint8
type BlueColor uint8
type WhiteColor uint8
type BlackColor uint8

// RGBA 实现 color.Color 接口
func (re RedColor) RGBA() (r, g, b, a uint32) {
	a = uint32(re) + 1
	a = a << 8
	return 0xFFFF, 0, 0, a - 1
}

// RGBA 实现 color.Color 接口
func (gr GreenColor) RGBA() (r, g, b, a uint32) {
	a = uint32(gr) + 1
	a = a << 8
	return 0, 0xFFFF, 0, a - 1
}

// RGBA 实现 color.Color 接口
func (bl BlueColor) RGBA() (r, g, b, a uint32) {
	a = uint32(bl) + 1
	a = a << 8
	return 0, 0, 0xFFFF, a - 1
}

// RGBA 实现 color.Color 接口
func (wc WhiteColor) RGBA() (r, g, b, a uint32) {
	a = uint32(wc) + 1
	a = a << 8
	return 0xFFFF, 0xFFFF, 0xFFFF, a - 1
}

// RGBA 实现 color.Color 接口
func (bc BlackColor) RGBA() (r, g, b, a uint32) {
	a = uint32(bc) + 1
	a = a << 8
	return 0, 0, 0, a - 1
}
