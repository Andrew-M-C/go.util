package draw

// Red 表示简单红色
const Red = RedColor(0xFF)

// Green 表示简单绿色
const Green = GreenColor(0xFF)

// Blue 表示简单绿色
const Blue = BlueColor(0xFF)

type RedColor uint8
type GreenColor uint8
type BlueColor uint8

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
