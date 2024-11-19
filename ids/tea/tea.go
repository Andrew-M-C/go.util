// Package tea 通过 TEA 算法来简单实现 8 位自增 id 类型的转换
package tea

import (
	"encoding/binary"

	"golang.org/x/crypto/tea"
)

// Key 定义一个加解密 key 大小
type Key [tea.KeySize]byte

// Encrypt 加密
func Encrypt(id uint64, key Key) uint64 {
	ci, _ := tea.NewCipher(key[:])

	// 申请一次空间就够了
	buffer := [tea.BlockSize * 2]byte{}
	in := buffer[:tea.BlockSize]
	out := buffer[tea.BlockSize:]

	binary.BigEndian.PutUint64(in, id)
	ci.Encrypt(out, in)

	return binary.BigEndian.Uint64(out)
}

// Decrypt 解密
func Decrypt(encrypted uint64, key Key) uint64 {
	ci, _ := tea.NewCipher(key[:])

	// 申请一次空间就够了
	buffer := [tea.BlockSize * 2]byte{}
	in := buffer[:tea.BlockSize]
	out := buffer[tea.BlockSize:]

	binary.BigEndian.PutUint64(in, encrypted)
	ci.Decrypt(out, in)

	return binary.BigEndian.Uint64(out)
}
