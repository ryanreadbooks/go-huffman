package huffman

import "fmt"

// BitsReader 定义比特的读取方式
type BitsReader struct {
	buf    []byte
	table  HuffmanDecTable
	index  int
	cursor uint64
	remain uint64
}

var (
	ErrBitCodeNotFound = fmt.Errorf("bitcode not found")
	ErrBitsExhausted = fmt.Errorf("bit exhausted")
)

func NewBitsReader(buf []byte, bitLen uint64, decodeTable HuffmanDecTable) *BitsReader {
	return &BitsReader{
		buf:    buf,
		table:  decodeTable,
		index:  0,
		cursor: 0,
		remain: bitLen,
	}
}

// ReadByte 从比特中解析出一个字节
func (r *BitsReader) ReadByte() (byte, error) {
	if r.remain == 0 {
		return 0, ErrBitsExhausted
	}

	// 从第index个字节的第cursor个比特开始读
	parsedCode := HuffmanCode{}
	var ret byte = 0

	i := 0
	foundOne := false

	// 每次最多读 MaxHuffmanCodeBitLen bit
	for ; i < MaxHuffmanCodeBitLen && r.remain > 0; i++ {
		if r.nextBit() {
			parsedCode.AppendOne()
		} else {
			parsedCode.AppendZero()
		}

		r.cursor = (r.cursor + 1) % 8
		r.remain--
		if r.cursor == 0 {
			r.index++
		}

		// 在decode table中找是否存在这个编码
		key, ok := r.table.Get(parsedCode)
		if ok {
			ret = key
			foundOne = true
			break
		}
	}

	if i == MaxHuffmanCodeBitLen {
		// 发现了不存在的比特编码
		return 0, ErrBitCodeNotFound
	}

	if r.remain == 0 && !foundOne {
		return 0, ErrBitsExhausted
	}

	return ret, nil
}

// ReadAll 解析所有比特位
func (r *BitsReader) ReadAll() ([]byte, error) {
	approxLen := r.remain / 8
	ret := make([]byte, 0, approxLen)

	for r.remain > 0 {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		ret = append(ret, b)
	}

	return ret, nil
}

// 判断下一个比特位（第index字节的第cursor个比特
// 返回true表示比特1，返回false表示比特0
func (r *BitsReader) nextBit() bool {
	mask := byte(0x80 >> r.cursor)
	return (r.buf[r.index] & mask) == mask
}
