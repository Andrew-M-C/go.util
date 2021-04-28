package bytes

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	// SDump() line format
	sDumpLineFormat     = `-------- | 00 01 02 03 04 05 06 07 | 08 09 10 11 12 13 14 15 | 01234567 89ABCDEF`
	sDumpLineHexPartLen = len(sDumpLineFormat) - len(`| 01234567 89ABCDEF`)
)

const easciiStart = 0xA0

var eascii = []rune{
	' ', //	A0	160	不换行空格（NBSP）
	'¡', //	A1	161	倒感叹号
	'¢', //	A2	162	英分
	'£', //	A3	163	英镑
	'¤', //	A4	164	货币记号
	'¥', //	A5	165	人民币/日元
	'¦', //	A6	166	断竖线
	'§', //	A7	167	小节符
	'¨', //	A8	168	分音符（元音变音）
	'©', //	A9	169	著作权符
	'ª', //	AA	170	阴性序数记号
	'«', //	AB	171	左指双尖引号
	'¬', //	AC	172	非标记
	'-', //AD	173	选择性连接号（SHY）
	'®', //	AE	174	注册商标
	'¯', //	AF	175	长音符
	'°', //	B0	176	度
	'±', //	B1	177	正负号
	'²', //	B2	178	二次方号
	'³', //	B3	179	三次方号
	'´', //	B4	180	锐音符
	'µ', //	B5	181	微符
	'¶', //	B6	182	段落标记
	'·', //	B7	183	中心点
	'¸', //	B8	184	软音符
	'¹', //	B9	185	一次方号
	'º', //	BA	186	阳性序数记号
	'»', //	BB	187	右指双尖引号
	'¼', //	BC	188	四分之一
	'½', //	BD	189	二分之一
	'¾', //	BE	190	四分之三
	'¿', //	BF	191	竖翻问号
	'À', //	C0	192	带抑音符的A
	'Á', //	C1	193	带锐音符的A
	'Â', //	C2	194	带扬抑符的A
	'Ã', //	C3	195	带颚化符的A
	'Ä', //	C4	196	带分音符的A
	'Å', //	C5	197	带上圆圈的A
	'Æ', //	C6	198	大写连字AE
	'Ç', //	C7	199	带下加符的C
	'È', //	C8	200	带抑音符的E
	'É', //	C9	201	带锐音符的E
	'Ê', //	CA	202	带扬抑符的E
	'Ë', //	CB	203	带分音符的E
	'Ì', //	CC	204	带抑音符的I
	'Í', //	CD	205	带锐音符的I
	'Î', //	CE	206	带扬抑符的I
	'Ï', //	CF	207	带分音符的I
	'Ð', //	D0	208	带横线符的D
	'Ñ', //	D1	209	带颚化符的N
	'Ò', //	D2	210	带抑音符的O
	'Ó', //	D3	211	带锐音符的O
	'Ô', //	D4	212	带扬抑符的O
	'Õ', //	D5	213	带颚化符的O
	'Ö', //	D6	214	带分音符的O
	'×', //	D7	215	乘号
	'Ø', //	D8	216	带斜线的O
	'Ù', //	D9	217	带抑音符的U
	'Ú', //	DA	218	带锐音符的U
	'Û', //	DB	219	带扬抑符的U
	'Ü', //	DC	220	带分音符的U
	'Ý', //	DD	221	带锐音符的Y
	'Þ', //	DE	222	清音p
	'ß', //	DF	223	清音s
	'à', //	E0	224	带抑音符的a
	'á', //	E1	225	带锐音符的a
	'â', //	E2	226	带扬抑符的a
	'ã', //	E3	227	带颚化符的a
	'ä', //	E4	228	带分音符的a
	'å', //	E5	229	带上圆圈的a
	'æ', //	E6	230	小写连字AE
	'ç', //	E7	231	带下加符的c
	'è', //	E8	232	带抑音符的e
	'é', //	E9	233	带锐音符的e
	'ê', //	EA	234	带扬抑符的e
	'ë', //	EB	235	带分音符的e
	'ì', //	EC	236	带抑音符的i
	'í', //	ED	237	带锐音符的i
	'î', //	EE	238	带扬抑符的i
	'ï', //	EF	239	带分音符的i
	'ð', //	F0	240	带斜线的d
	'ñ', //	F1	241	带颚化符的n
	'ò', //	F2	242	带抑音符的o
	'ó', //	F3	243	带锐音符的o
	'ô', //	F4	244	带扬抑符的o
	'õ', //	F5	245	带颚化符的o
	'ö', //	F6	246	带分音符的o
	'÷', //	F7	247	除号
	'ø', //	F8	248	带斜线的o
	'ù', //	F9	249	带抑音符的u
	'ú', //	FA	250	带锐音符的u
	'û', //	FB	251	带扬抑符的u
	'ü', //	FC	252	带分音符的u
	'ý', //	FD	253	带锐音符的y
	'þ', //	FE	254	小写字母Thorn
	'ÿ', //	FF	255	带分音符的y
}

// SDump dump s byte slice into string.
// string format: "00000000 | 00 11 22 33 44 55 66 77 | 88 99 AA BB CC DD EE FF | 01234567 89ABCDEF"
func SDump(b []byte, desc ...string) string {
	le := len(b)
	buff := bytes.Buffer{}

	if len(desc) > 0 && desc[0] != "" {
		buff.WriteString(fmt.Sprintf("byte slice %s, length %d", desc[0], le))
	} else {
		buff.WriteString(fmt.Sprintf("byte slice %p, length %d", b, le))
	}

	buff.WriteRune('\n')
	buff.WriteString(sDumpLineFormat)

	const lineCount = 16
	lineHexBuff := bytes.Buffer{}
	lineChrBuff := bytes.Buffer{}

	runeByByte := func(chr byte) rune {
		if chr < 0x20 {
			return '.'
		} else if chr < 0x7F {
			return rune(chr)
		} else if chr < easciiStart {
			return '.'
		} else {
			return eascii[int(chr)-easciiStart]
		}
	}

	offset := 0
	for ; offset+lineCount < le; offset += lineCount {
		lineHexBuff.WriteString(fmt.Sprintf("%08X | ", offset))
		lineChrBuff.WriteString("| ")

		for i := 0; i < lineCount; i++ {
			if i == 8 {
				lineHexBuff.WriteString("| ")
				lineChrBuff.WriteRune(' ')
			}
			chr := b[offset+i]
			lineHexBuff.WriteString(fmt.Sprintf("%02X ", chr))
			lineChrBuff.WriteRune(runeByByte(chr))
		}

		buff.WriteRune('\n')
		buff.Write(lineHexBuff.Bytes())
		buff.Write(lineChrBuff.Bytes())
		lineHexBuff.Reset()
		lineChrBuff.Reset()
	}

	// dump remaining line
	if offset >= le {
		return buff.String()
	}

	lineHexBuff.WriteString(fmt.Sprintf("%08X | ", offset))
	lineChrBuff.WriteString("| ")
	for i := 0; offset+i < le; i++ {
		if i == 8 {
			lineHexBuff.WriteString("| ")
			lineChrBuff.WriteRune(' ')
		}
		chr := b[offset+i]
		lineHexBuff.WriteString(fmt.Sprintf("%02X ", chr))
		lineChrBuff.WriteRune(runeByByte(chr))
	}

	lack := sDumpLineHexPartLen - lineHexBuff.Len()
	lineHexBuff.WriteString(strings.Repeat(" ", lack))

	buff.WriteRune('\n')
	buff.Write(lineHexBuff.Bytes())
	buff.Write(lineChrBuff.Bytes())

	return buff.String()
}
