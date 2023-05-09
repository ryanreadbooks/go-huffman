package huffman

import (
	"fmt"
	"strings"
)

type Frequencies map[byte]uint64

func (f Frequencies) Increment(key byte) {
	f[key]++
}

type HuffmanCodeInterface interface {
	fmt.Stringer
	BitLen() int
	AppendOne()
	AppendZero()
	ReverseNew() HuffmanCodeInterface
	Clone() HuffmanCodeInterface
}

const (
	MaxHuffmanCodeBitLen = 24 // 允许最长的编码比特位到24位
)

// 用一个uint32类型代表一个huffman的二进制编码格式
// 高8位存比特位长度，低24位存比特位本身
// 比特位本身的最高位放在Uint32中低24位的最高位
type HuffmanCode struct {
	bits uint32
}

// NewHuffmanCodeFromString 从01字符串中创建对象
func NewHuffmanCodeFromString(s string) *HuffmanCode {
	code := &HuffmanCode{}

	for _, ch := range s {
		if ch == '0' {
			code.AppendZero()
		} else if ch == '1' {
			code.AppendOne()
		} else {
			panic(fmt.Sprintf("character must be either '0' or '1', but found %c", ch))
		}
	}

	return code
}

// BitLen 返回比特位长度
func (h *HuffmanCode) BitLen() int {
	// 高8位
	return int(uint32(h.bits&0xFF000000) >> 24)
}

func (h *HuffmanCode) setBitLen(l uint8) {
	high := uint32(l) << 24
	h.bits &= 0x00FFFFFF
	h.bits |= high
}

func (h *HuffmanCode) BitsLow16() uint16 {
	return uint16(h.bits & 0x0000FFFF)
}

// Bits 以uint32的形式返回比特位本身
// 丢掉高8位的长度 并将有效比特位的最高位移到从uint32最高位开始
func (h *HuffmanCode) Bits() uint32 {
	return (h.bits & 0x00FFFFFF) << 8
}

// Bits 以uint32的形式返回比特位本身
// 但是不进行移位
func (h *HuffmanCode) BitsUntouched() uint32 {
	return (h.bits & 0x00FFFFFF)
}

// AllBits 以uint32的形式返回带有bitlen的比特位
func (h *HuffmanCode) AllBits() uint32 {
	return h.bits
}

// 实现fmt.Stringer接口
func (h *HuffmanCode) String() string {
	bitLen := h.BitLen()
	res := strings.Builder{}
	res.Grow(bitLen)
	bits := h.Bits()
	var mask uint32 = 0x80000000

	for i := 0; i < bitLen; i++ {
		if (mask>>i)&bits == 0 {
			res.WriteByte('0')
		} else {
			res.WriteByte('1')
		}
	}

	return res.String()
}

// AppendOne 往比特位后追加1
func (h *HuffmanCode) AppendOne() {
	oldBitLen := h.BitLen()
	if oldBitLen >= MaxHuffmanCodeBitLen {
		fmt.Printf("someone try to append more bit: %s\n", h.String())
		return
	}

	shift := 23 - oldBitLen
	h.bits |= uint32(1 << shift)

	h.setBitLen(uint8(oldBitLen) + 1)
}

// AppendOne 往比特位后追加0
func (h *HuffmanCode) AppendZero() {
	oldBitLen := h.BitLen()
	if oldBitLen >= MaxHuffmanCodeBitLen {
		return
	}

	h.setBitLen(uint8(oldBitLen) + 1)
}

// ReverseNew 逆序比特位并返回新的对象
func (h *HuffmanCode) ReverseNew() *HuffmanCode {
	clone := &HuffmanCode{}
	bitLen := h.BitLen()
	if bitLen == 0 {
		return clone
	}

	bits := h.BitsUntouched()
	bits >>= (24 - bitLen)
	var mask uint32 = 0x1

	for i := 0; i < bitLen; i++ {
		if (bits>>i)&mask == 0 {
			clone.AppendZero()
		} else {
			clone.AppendOne()
		}
	}

	return clone
}

// Clone 返回对象的拷贝
func (h *HuffmanCode) Clone() *HuffmanCode {
	return &HuffmanCode{
		bits: h.bits,
	}
}
