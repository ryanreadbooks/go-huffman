package huffman

import (
	"errors"
)

const (
	MaxUint16Len = 16
	MaxUint32Len = 32
)

var (
	ErrMaxLenExceeded = errors.New("maximum length exceeded")
)

// BitsWriter 定义了比特的写入方式
type BitsWriter struct {
	// 存放比特流的缓冲区
	buf []byte
	// 当前字节索引
	idx uint64
	// 当前比特索引
	slot uint64
}

func NewBitsWriter() *BitsWriter {
	return &BitsWriter{
		buf:  make([]byte, 1),
		idx:  0,
		slot: 0,
	}
}

func (w *BitsWriter) updateCursor() {
	w.slot = (w.slot + 1) % 8
	if w.slot == 0 {
		w.idx++
		w.buf = append(w.buf, 0)
	}
}

func (w *BitsWriter) appendOne() {
	mask := uint8(1 << (7 - w.slot))
	w.buf[w.idx] |= mask
	w.updateCursor()
}

func (w *BitsWriter) appendZero() {
	w.updateCursor()
}

// WriteUint16 将n位比特写入，从高位开始写，最多写入16位
func (w *BitsWriter) WriteUint16(a uint16, n uint8) error {
	if n > MaxUint16Len {
		return ErrMaxLenExceeded
	}

	// 特殊情况处理
	if w.slot == 0 && (n == 8 || n == 16) {
		high := byte((a & 0xFF00) >> 8)
		if n == 8 {
			w.buf = append(w.buf, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
		} else if n == 16 {
			low := byte(a & 0x00FF)
			w.buf = append(w.buf, 0, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
			w.buf[w.idx] = low
			w.idx += 1
		}
		return nil
	}

	// slot != 0 or n % 8 != 0
	var mask uint16 = 0x8000
	for i := 0; i < int(n); i++ {
		if (a<<i)&mask == 0 {
			w.appendZero()
		} else {
			w.appendOne()
		}
	}

	return nil
}

// WriteUint32 将n位比特写入，从高位开始写，最多写入32位
func (w *BitsWriter) WriteUint32(a uint32, n uint8) error {
	if n > MaxUint32Len {
		return ErrMaxLenExceeded
	}

	// 特殊情况处理，可能有刚好是1、2、3、4个字节
	if w.slot == 0 && (n == 8 || n == 16 || n == 24 || n == 32) {
		high := byte((a & 0xFF000000) >> 24) // 32位的高八位
		if n == 8 {
			w.buf = append(w.buf, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
		} else if n == 16 {
			middleLeft := byte((a & 0x00FF0000) >> 16)
			w.buf = append(w.buf, 0, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
			w.buf[w.idx] = middleLeft
			w.idx += 1
		} else if n == 24 {
			middleLeft := byte((a & 0x00FF0000) >> 16)
			middleRight := byte((a & 0x0000FF00) >> 8)
			w.buf = append(w.buf, 0, 0, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
			w.buf[w.idx] = middleLeft
			w.idx += 1
			w.buf[w.idx] = middleRight
			w.idx += 1
		} else { // n == 32
			middleLeft := byte((a & 0x00FF0000) >> 16)
			middleRight := byte((a & 0x0000FF00) >> 8)
			low := byte(a & 0x000000FF)
			w.buf = append(w.buf, 0, 0, 0, 0, 0)
			w.buf[w.idx] = high
			w.idx += 1
			w.buf[w.idx] = middleLeft
			w.idx += 1
			w.buf[w.idx] = middleRight
			w.idx += 1
			w.buf[w.idx] = low
			w.idx += 1
		}
		return nil
	}

	// 一般情况，写入n个比特
	var mask uint32 = 0x80000000
	for i := 0; i < int(n); i++ {
		if (a<<i)&mask == 0 {
			w.appendZero()
		} else {
			w.appendOne()
		}
	}

	return nil
}

// Buf 返回底层比特位缓冲区的拷贝
func (w *BitsWriter) Buf() []byte {
	// 这里仅需要拷贝有效的数据，而不是整个w.buf缓冲区，否则的话数据中会包含大量的无效0
	var validByteLen int
	if w.slot == 0 {
		validByteLen = int(w.idx)
	} else {
		validByteLen = int(w.idx) + 1
	}
	cp := make([]byte, validByteLen)
	copy(cp, w.buf)

	return cp
}
