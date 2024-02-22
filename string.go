package emozi

import (
	"errors"
	"hash/crc32"
	"strings"

	base14 "github.com/fumiama/go-base16384"
)

var (
	ErrInvalidEmoziString = errors.New("invalid EmoziString")
)

// EmoziString 一个颜文字汉字转写串, 包含串头 校验字节数*2 字节校验和
type EmoziString string

// WrapRawEmoziString 为不包含串头的转写串加一个头使其成为合法 EmoziString
func WrapRawEmoziString(s string) EmoziString {
	rs := []rune(s)
	if len(rs) < 4 {
		diff := 4 - len(rs)
		for i := 0; i < diff; i++ {
			rs = append(rs, EmptyMark)
		}
		s = string(rs)
	}
	h := crc32.NewIEEE()
	h.Write(base14.StringToBytes(s))
	sum := h.Sum32() % 校验模
	sb := strings.Builder{}
	buf := [校验字节数]uint8{}
	for i := 校验字节数 - 1; i >= 0; i-- {
		buf[i] = uint8((sum / 校验倍数[i]) % uint32(校验表长度))
	}
	for i, n := range buf {
		sb.WriteRune(rs[i])
		sb.WriteRune(校验表[n])
	}
	sb.WriteString(string(rs[len(buf):]))
	return EmoziString(sb.String())
}

// String 输出不包含串头的转写串
func (es EmoziString) String() string {
	if !es.IsValid() {
		return ErrInvalidEmoziString.Error()
	}
	rs := []rune(es)
	sb := strings.Builder{}
	for i := 0; i < 校验字节数; i++ {
		sb.WriteRune(rs[i*2])
	}
	sb.WriteString(string(rs[校验字节数*2:]))
	return sb.String()
}

// IsValid 判断是否是合法 EmoziString
func (es EmoziString) IsValid() bool {
	rs := []rune(es)
	if len(rs) < 校验字节数+4 {
		return false
	}
	h := crc32.NewIEEE()
	sum := uint32(0)
	for i := 0; i < 校验字节数; i++ {
		h.Write(base14.StringToBytes(string(rs[i*2])))
		sum += 校验倍数[i] * uint32(逆校验表[rs[i*2+1]])
	}
	h.Write(base14.StringToBytes(string(rs[校验字节数*2:])))
	return h.Sum32()%校验模 == sum
}
