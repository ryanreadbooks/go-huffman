package huffman

import (
	"fmt"
	"hash/crc32"
	"log"
)

type huffmanDeserializer interface {
	ItemNum() int
}

// Huffman编码码表
type HuffmanEncTable map[byte]*HuffmanCode

// Huffman解码码表
type HuffmanDecTable map[HuffmanCode]byte

const (
	HuffmanEncTableSerStartFlag uint32 = 0x48464553 // "HFES"
	HuffmanEncTableSerEndFlag   uint32 = 0x48464545 // "HFEE"
	ChecksumPoly                       = 0xD5828281
)

const (
	MetaSize               = 4            // flag or numbder size
	MinHuffmanTableSerSize = 4 * MetaSize // bytes

	Uint32ByteSize = 4 // bytes
	TableItemSize  = 5 // bytes
)

var (
	ErrInvalidSize        = fmt.Errorf(fmt.Sprintf("len of data is small then %d", MinHuffmanTableSerSize))
	ErrInvalidStartFlag   = fmt.Errorf("start flag invalid")
	ErrInvalidEndFlag     = fmt.Errorf("end flag invalid")
	ErrCursorOverflow     = fmt.Errorf("cursor overflow")
	ErrChecksumNotMatched = fmt.Errorf("checksum not matched")
	ErrDeserialize        = fmt.Errorf("parse error")
)

var (
	crc32q = crc32.MakeTable(ChecksumPoly)
)

func NewHuffmanEncTable(tree *HuffmanTree) HuffmanEncTable {
	table := make(HuffmanEncTable, len(tree.Leaves))
	for _, leaf := range tree.Leaves {
		table[leaf.Byte] = leaf.Code
	}

	return table
}

// Get 获取HuffmanEncTable中的编码
func (h HuffmanEncTable) Get(key byte) *HuffmanCode {
	if node, ok := h[key]; ok {
		return node
	}
	return nil
}

// ItemNum 返回HuffmanEncTable中表项的数量
func (h HuffmanEncTable) ItemNum() int {
	return len(h)
}

// Equals 判断两个HuffmanEncTable是否相同
func (h HuffmanEncTable) Equals(other HuffmanEncTable) bool {
	if len(h) != len(other) {
		return false
	}
	for k, v := range h {
		ov, ok := other[k]
		if !ok {
			return false
		}
		if ov.Bits() != v.Bits() {
			return false
		}
	}
	return true
}

func NewHuffmanDecTable(n int) HuffmanDecTable {
	return make(HuffmanDecTable, n)
}

// Get 获取某个编码所对应的字节
func (h HuffmanDecTable) Get(key HuffmanCode) (byte, bool) {
	v, ok := h[key]
	return v, ok
}

// ItemNum 返回HuffmanDecTable中表项的数量
func (h HuffmanDecTable) ItemNum() int {
	return len(h)
}

// Serialize 将HuffmanEncTable序列化到字节切片中
// 序列化格式如下（大端序：高位放在低地址，低位放在高地址）
// START_FLAG				4 bytes
// NUMBER OF TABLE ITEMS	4 bytes uint32
// TABLE_ITEM_1(BYTE+CODE)	1+4=5 bytes
// TABLE_ITEM_2(BYTE+CODE)	1+4=5 bytes
// ...
// TABLE_ITEM_N(BYTE+CODE)	1+4=5 bytes
// CRC32					4 bytes
// END_FLAG					4 bytes
func (h HuffmanEncTable) Serialize() ([]byte, error) {
	n := len(h)
	size := MinHuffmanTableSerSize + 5*n
	ser := make([]byte, 0, size)

	// 写入开始标志
	ser = writeUint32ToBytes(HuffmanEncTableSerStartFlag, ser)
	// 写入数量
	ser = writeUint32ToBytes(uint32(n), ser)
	// 写入表项
	for key, code := range h {
		ser = append(ser, key)
		ser = writeUint32ToBytes(code.AllBits(), ser)
	}
	// 写入前面内容的校验和
	checksum := crc32.Checksum(ser, crc32q)
	ser = writeUint32ToBytes(checksum, ser)

	// 写入结束标志
	ser = writeUint32ToBytes(HuffmanEncTableSerEndFlag, ser)
	return ser, nil
}

func parseFlag(data []byte, cursor int, flag uint32) (int, error) {
	// 读标志
	got, err := readNextUint32(data, cursor)
	if err != nil {
		return 0, err
	}
	if got != flag {
		return 0, fmt.Errorf("wrong flag")
	}
	cursor += MetaSize

	return cursor, nil
}

func parseStartFlag(data []byte, cursor int) (int, error) {
	// 读开始标志
	cursor, err := parseFlag(data, cursor, HuffmanEncTableSerStartFlag)
	if err != nil {
		return 0, ErrInvalidStartFlag
	}

	return cursor, nil
}

func parseEndFlag(data []byte, cursor int) (int, error) {
	// 读开始标志
	cursor, err := parseFlag(data, cursor, HuffmanEncTableSerEndFlag)
	if err != nil {
		return 0, ErrInvalidEndFlag
	}

	return cursor, nil
}

func parseItemNum(data []byte, cursor int) (int, int, error) {
	// 读表项的数量
	itemNum, err := readNextUint32(data, cursor)
	if err != nil {
		return 0, 0, err
	}
	cursor += MetaSize

	return int(itemNum), cursor, nil
}

func parseEncTable(data []byte, cursor int, itemNum int) (HuffmanEncTable, int, error) {
	retHuff := make(HuffmanEncTable, itemNum)

	for i := 0; i < int(itemNum); i++ {
		key, code, err := readNextTableItem(data, cursor)
		if err != nil {
			return nil, 0, err
		}
		cursor += TableItemSize
		retHuff[key] = &HuffmanCode{bits: code}
	}

	return retHuff, cursor, nil
}

func validateChecksum(data []byte, cursor int) (int, error) {
	expectedChecksum, err := readNextUint32(data, cursor) // 数据中的已有的checksum
	if err != nil {
		return 0, err
	}
	// 计算cursor前面的checksum
	calChecksum := crc32.Checksum(data[0:cursor], crc32q)
	if expectedChecksum != calChecksum {
		log.Printf("expected checksum is %x, but got %x\n", expectedChecksum, calChecksum)
		return 0, ErrChecksumNotMatched
	}
	cursor += MetaSize

	return cursor, nil
}

func parseDecTable(data []byte, cursor int, itemNum int) (HuffmanDecTable, int, error) {
	retHuff := make(HuffmanDecTable, itemNum)

	for i := 0; i < int(itemNum); i++ {
		key, code, err := readNextTableItem(data, cursor)
		if err != nil {
			return nil, 0, err
		}
		cursor += TableItemSize
		retHuff[HuffmanCode{bits: code}] = key
	}

	return retHuff, cursor, nil
}

type huffTableItemParser func(data []byte, cursor int, itemNum int) (huffmanDeserializer, int, error)

// 反序列化字节切片
func deserialize(data []byte, parser huffTableItemParser) (huffmanDeserializer, error) {
	n := len(data)
	if n < MinHuffmanTableSerSize {
		return nil, ErrInvalidSize
	}

	// 读开始标志
	cursor := 0
	cursor, err := parseStartFlag(data, cursor)
	if err != nil {
		return nil, err
	}

	// 读表项的数量
	itemNum, cursor, err := parseItemNum(data, cursor)
	if err != nil {
		return nil, err
	}

	// 解析表项内容
	huffTable, cursor, err := parser(data, cursor, itemNum)
	if err != nil {
		return nil, err
	}

	// 检查检验和
	cursor, err = validateChecksum(data, cursor)
	if err != nil {
		return nil, err
	}

	// 检查结束标志
	_, err = parseEndFlag(data, cursor)
	if err != nil {
		return nil, err
	}

	return huffTable, nil
}

// DeserializeHuffmanEncTable 将字节切片反序列回HuffmanEncTable
func DeserializeHuffmanEncTable(data []byte) (HuffmanEncTable, error) {
	huffTable, err := deserialize(data, func(data []byte, cursor, itemNum int) (huffmanDeserializer, int, error) {
		return parseEncTable(data, cursor, itemNum)
	})

	if err != nil {
		return nil, err
	}

	if encTable, ok := huffTable.(HuffmanEncTable); ok {
		return encTable, nil
	}
	return nil, ErrDeserialize
}

// DeserializeHuffmanDecTable 将字节切片反序列回HuffmanDecTable
func DeserializeHuffmanDecTable(data []byte) (HuffmanDecTable, error) {
	huffTable, err := deserialize(data, func(data []byte, cursor, itemNum int) (huffmanDeserializer, int, error) {
		return parseDecTable(data, cursor, itemNum)
	})

	if err != nil {
		return nil, err
	}

	if decTable, ok := huffTable.(HuffmanDecTable); ok {
		return decTable, nil
	}
	return nil, ErrDeserialize
}

func readNextUint32(buf []byte, start int) (uint32, error) {
	n := len(buf)
	if n < Uint32ByteSize {
		return 0, ErrInvalidSize
	}
	if start+3 > n {
		return 0, ErrCursorOverflow
	}

	byte0 := buf[start]
	byte1 := buf[start+1]
	byte2 := buf[start+2]
	byte3 := buf[start+3]

	var ans uint32 = 0
	ans |= uint32(byte0) << 24
	ans |= uint32(byte1) << 16
	ans |= uint32(byte2) << 8
	ans |= uint32(byte3)

	return ans, nil
}

func readNextTableItem(buf []byte, start int) (byte, uint32, error) {
	n := len(buf)
	if n < TableItemSize {
		return 0, 0, ErrInvalidSize
	}
	if start+4 > n {
		return 0, 0, ErrCursorOverflow
	}

	key := buf[start]
	code, err := readNextUint32(buf, start+1)
	if err != nil {
		return 0, 0, err
	}
	return key, code, nil
}
