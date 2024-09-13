// Package residentid 实现居民身份证号的处理逻辑
package residentid

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Andrew-M-C/go.util/china/admindivision"
)

var beijing = time.FixedZone("Asia/Beijing", 8*60*60)

// Gender 性别, 只分男女
type Gender int8

const (
	// UnknownGender 只在 ID 不合法的时候返回
	UnknownGender Gender = iota
	// Male 男性
	Male
	// Female 女性
	Female
)

func (g Gender) String() string {
	switch g {
	case Male:
		return "男性"
	case Female:
		return "女性"
	default:
		return "未知性别"
	}
}

// DetailInfo 表示身份证详细信息, 便于一次性获取
type DetailInfo struct {
	Hometown []*admindivision.Division
	Birthday time.Time
	Gender   Gender
}

func (inf DetailInfo) String() string {
	return fmt.Sprintf(
		"{性别: %v, 出生: %s, 生日: %s}",
		inf.Gender,
		admindivision.DescribeDivisionChain(inf.Hometown, ""),
		inf.Birthday.Format("2006/01/02"),
	)
}

// ID 表示一个身份证号及相关信息
type ID struct {
	raw      string
	birthday time.Time
}

// New 新建一个 ID
func New(num string) (ID, error) {
	if err := validateDigit(num); err != nil {
		return ID{}, err
	}
	if err := validateChecksum(num); err != nil {
		return ID{}, err
	}
	birthday, err := validateBirthday(num)
	if err != nil {
		return ID{}, err
	}
	id := ID{
		raw:      num,
		birthday: birthday,
	}
	return id, nil
}

// DetailInfo 返回详细描述
func (id ID) DetailInfo() DetailInfo {
	if id.raw == "" {
		return DetailInfo{}
	}
	inf := DetailInfo{
		Hometown: id.Hometown(),
		Birthday: id.Birthday(),
		Gender:   id.Gender(),
	}
	return inf
}

// Hometown 返回身份证上的家乡信息
func (id ID) Hometown() []*admindivision.Division {
	if id.raw == "" {
		return nil
	}
	locationCode := id.raw[:6]
	return admindivision.SearchDivisionByCode(locationCode)
}

// Birthday 生日
func (id ID) Birthday() time.Time {
	return id.birthday
}

// BirthSequence 返回在同行政区划中同性别的身份证标记序号。序号一以 0 表示, 非法返回 -1。
func (id ID) BirthSequence() int {
	if id.raw == "" {
		return -1
	}
	s := id.raw[14 : 14+3]
	u, _ := strconv.ParseUint(s, 10, 32)
	return int(u / 2)
}

// Gender 性别
func (id ID) Gender() Gender {
	if id.raw == "" {
		return UnknownGender
	}
	switch id.raw[16] {
	case '1', '3', '5', '7', '9':
		return Male
	default:
		return Female
	}
}

func (id ID) String() string {
	if id.raw == "" {
		return "<no ID>"
	}
	return id.raw
}

func validateDigit(num string) error {
	if len(num) != 18 {
		return errors.New("invalid length")
	}
	for i, r := range num {
		if r >= '0' && r <= '9' {
			continue
		}
		// 注意一下，这两个不同，一个是字母X，另一个是罗马数字
		if i == 17 && isTen(r) {
			continue
		}
		return fmt.Errorf("invalid character '%c'", r)
	}

	return nil
}

func validateBirthday(num string) (time.Time, error) {
	const layout = "20060102"
	s := num[6 : 6+8]
	tm, err := time.ParseInLocation(layout, s, beijing)
	if err != nil {
		return time.Time{}, err
	}

	if tm.Year() <= 1800 {
		return tm, errors.New("invalid year")
	}
	return tm, nil
}

// reference: [居民身份证查询验证](http://www.ip33.com/shenfenzheng.html)
func validateChecksum(num string) error {
	coefficients := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2} // 总共17项，表示 b0 ~ b16 的系数
	remainVal := []uint8{1, 0, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	digits, err := convIDToInts(num)
	if err != nil {
		return err
	}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(digits[i]) * coefficients[i]
	}

	remain := sum % 11
	if digits[17] != remainVal[remain] {
		return errors.New("checksum failed")
	}

	return nil
}

func convIDToInts(s string) ([]uint8, error) {
	ret := make([]uint8, 18)
	for i, r := range s {
		if r >= '0' && r <= '9' {
			ret[i] = uint8(r - '0')
			continue
		}
		// 注意一下，这两个不同，一个是字母X，另一个是罗马数字
		if i == 17 && isTen(r) {
			ret[17] = 10
			continue
		}
		return nil, errors.New("checksum failed")
	}
	return ret, nil
}

func isTen(r rune) bool {
	return r == 'X' || r == 'Ⅹ'
}
