package huffman

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path"
)

const (
	CompressedFileStartFlag uint16 = 0x5259
	CompressedFileEndFlag   uint16 = 0x414E
)

var (
	ErrCanNotParseFileHeader = fmt.Errorf("can not parse file header")
)

// compressBytesWith 使用给定的Huffman编码表压缩字节切片
// 返回压缩后的字节切片，压缩后的有效比特数
func compressBytesWith(data []byte, table HuffmanEncTable) ([]byte, uint64, error) {
	w := NewBitsWriter()
	var totalBits uint64 = 0

	// 遍历data，编码每个字节
	for i, b := range data {
		code := table.Get(b)
		if code == nil {
			return nil, 0, fmt.Errorf("code for %b(%c at %d) not found", b, b, i)
		}
		bitlen := code.BitLen()
		totalBits += uint64(bitlen)
		err := w.WriteUint32(code.Bits(), uint8(bitlen))
		if err != nil {
			return nil, 0, fmt.Errorf(err.Error())
		}
	}
	return w.Buf(), totalBits, nil
}

// CompressBytes 压缩一个字节切片
// 返回压缩后的字节切片，压缩后的有效比特数
func CompressBytes(data []byte) ([]byte, uint64, error) {
	// 计算各个字节出现的频数
	freq := CountFrequencies(data)
	// 构建Huffman树
	tree := NewHuffmanTree(freq)
	// 获取Huffman编码表
	table := NewHuffmanEncTable(tree)

	return compressBytesWith(data, table)
}

// CompressFile 压缩一个文件
// 将src文件压缩，然后写入到dst文件中
//
// 压缩文件格式如下：（大端序）
// HEADER
//   - START_FLAG						2 bytes (uint16)
//   - SRC_FILENAME_LEN					2 bytes (uint16)
//   - BYTE SIZE BEFORE COMPRESSION		4 bytes (uint32)
//   - BYTE SIZE AFTER COMPRESSION		4 bytes (uint32)
//   - SRC_FILENAME						n bytes
//
// DATA
//   - HUFFMAN TABLE
//     -- HUFFMAN TABLE SIZE 	4 bytes (uint32)
//     -- HUFFMAN TABLE DATA
//   - COMPRESSED DATA
//     -- VALID BIT LEN			4 bytes (uint32) + 1 bytes = 5 bytes
//     -- COMPRESSED BIT
//
// TAIL
//   - CRC32 CHECKSUM	  	4 bytes (uint32)
//   - END_FLAG				2 bytes (uint16)
func CompressFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	// 读入源文件内容
	allSrcBytes, err := io.ReadAll(srcF)
	if err != nil {
		return err
	}

	// 执行源文件压缩
	freq := CountFrequencies(allSrcBytes)
	tree := NewHuffmanTree(freq)
	encTable := NewHuffmanEncTable(tree)

	compressedBytes, bitLen, err := compressBytesWith(allSrcBytes, encTable)
	if err != nil {
		return err
	}

	// Huffman码表
	encTableSer, err := encTable.Serialize()
	if err != nil {
		return err
	}

	// 准备写入目标文件
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	originalSize := uint32(len(allSrcBytes))
	compressedSize := uint32(len(compressedBytes))
	filenameNoDir := path.Base(src)
	filenameNoDirSize := uint16(len(filenameNoDir))

	// 27字节为文件的固定开销
	var alloc uint32 = 27 + compressedSize + uint32(len(encTableSer))
	dstBytes := make([]byte, 0, alloc)
	// 写入文件头
	dstBytes = writeUint16ToBytes(CompressedFileStartFlag, dstBytes) // 文件开始标识
	dstBytes = writeUint16ToBytes(filenameNoDirSize, dstBytes)       // 文件名长度
	dstBytes = writeUint32ToBytes(originalSize, dstBytes)            // 压缩前字节大小
	dstBytes = writeUint32ToBytes(compressedSize, dstBytes)          // 压缩后字节大小
	dstBytes = append(dstBytes, []byte(filenameNoDir)...)            // 源文件名

	// 写入数据区
	dstBytes = writeUint32ToBytes(uint32(len(encTableSer)), dstBytes) // Huffman码表大小
	dstBytes = append(dstBytes, encTableSer...)                       // Huffman码表

	// 根据实际比特长度计算压缩后需要占用多少个字节
	bytesNeededAfterCompressed := bitLen / 8
	slot := bitLen % 8
	if slot != 0 {
		bytesNeededAfterCompressed += 1
	}

	// 用5个字节来记录bitLen：
	// bytesNeededAfterCompressed用4个字节
	// slot用1个字节
	dstBytes = writeUint32ToBytes(uint32(bytesNeededAfterCompressed), dstBytes)
	dstBytes = append(dstBytes, byte(slot))
	// dstBytes = writeUint64ToBytes(bitLen, dstBytes) // bitlen
	dstBytes = append(dstBytes, compressedBytes...) // 压缩后数据本身

	// 写入文件尾
	checksum := crc32.Checksum(dstBytes, crc32q)
	dstBytes = writeUint32ToBytes(checksum, dstBytes)              // 校验和
	dstBytes = writeUint16ToBytes(CompressedFileEndFlag, dstBytes) // 结束标记

	// 一次性写入文件
	n, err := dstF.Write(dstBytes)
	if err != nil {
		return err
	}

	log.Printf("successfully written %d bytes into %s\n", n, dst)

	return nil
}

func decompressBytesWith(data []byte, bitLen uint64, table HuffmanDecTable) ([]byte, error) {
	reader := NewBitsReader(data, bitLen, table)

	recovery, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return recovery, nil

}

// DecompressBytes 解压缩一个字节切片
// 输入参数包括压缩了的字节切片本身，字节切片中有效比特数和Huffman解码表
func DecompressBytes(data []byte, bitLen uint64, table HuffmanDecTable) ([]byte, error) {
	return decompressBytesWith(data, bitLen, table)
}

// DecompressFile 解压缩一个文件
// 将src文件解压缩，然后写入到dst文件中
func DecompressFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	srcBytes, err := io.ReadAll(srcF)
	if err != nil {
		return err
	}

	// 解析源文件的压缩了的字节
	cursor := 0
	// 文件头
	cursor, err = parseFileHeader(srcBytes, cursor)
	if err != nil {
		return fmt.Errorf("can not parse file header: %v", err)
	}

	// 数据区
	decompressedBytes, cursor, err := parseCompressedDataArea(srcBytes, cursor)
	if err != nil {
		return fmt.Errorf("can not parse file data area: %v", err)
	}

	// 文件尾
	// 校验数据是否正确
	_, err = parseFileTail(srcBytes, cursor)
	if err != nil {
		return fmt.Errorf("can not parse file tail: %v", err)
	}

	// 创建目标文件准备写回
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	n, err := dstF.Write(decompressedBytes)
	if err != nil {
		return err
	}
	log.Printf("successfully written %d bytes into destination: %s\n", n, dst)

	return nil
}

// 解析压缩文件头
func parseFileHeader(srcBytes []byte, cursor int) (newCursor int, err error) {
	defer func() {
		if p := recover(); p != nil {
			// 这里捕获可能的切片访问越界造成的panic
			newCursor = 0
			err = fmt.Errorf("%v", p)
		}
	}()

	// 文件开始标记
	gotStartFlag, err := readNextUint16(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	if gotStartFlag != CompressedFileStartFlag {
		return 0, ErrInvalidStartFlag
	}
	cursor += Uint16ByteSize

	// 压缩前文件名的长度
	beforeFilenameLen, err := readNextUint16(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	cursor += Uint16ByteSize

	// 32bit的压缩前文件大小
	_, err = readNextUint32(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	cursor += Uint32ByteSize

	// 32bit的压缩后文件大小
	_, err = readNextUint32(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	cursor += Uint32ByteSize

	// 源文件名字
	end := cursor + int(beforeFilenameLen)
	if end > len(srcBytes) {
		return 0, ErrCursorOverflow
	}
	_ = srcBytes[cursor:end]
	cursor = end

	return cursor, nil
}

// 解析压缩文件数据区
func parseCompressedDataArea(srcBytes []byte, cursor int) (data []byte, newCursor int, err error) {
	defer func() {
		if p := recover(); p != nil {
			// 这里捕获可能的切片访问越界造成的panic
			data = nil
			newCursor = 0
			err = fmt.Errorf("%v", p)
		}
	}()

	// Huffman码表
	huffTableLen, err := readNextUint32(srcBytes, cursor)
	if err != nil {
		return nil, 0, err
	}
	cursor += Uint32ByteSize

	decTable, err := DeserializeHuffmanDecTable(srcBytes[cursor : cursor+int(huffTableLen)])
	if err != nil {
		return nil, 0, err
	}
	cursor += int(huffTableLen)

	// 压缩数据解析
	compressedBytesLen, err := readNextUint32(srcBytes, cursor)
	if err != nil {
		return nil, 0, err
	}
	cursor += Uint32ByteSize
	slot := uint8(srcBytes[cursor])
	cursor += 1

	var validBitLen uint64
	if slot == 0 {
		validBitLen = uint64(compressedBytesLen * 8)
	} else {
		validBitLen = uint64((compressedBytesLen-1)*8 + uint32(slot))
	}

	decompressedBytes, err := decompressBytesWith(srcBytes[cursor:], validBitLen, decTable)
	if err != nil {
		return nil, 0, err
	}
	cursor += int(compressedBytesLen)

	return decompressedBytes, cursor, nil
}

// 解析压缩文件尾
func parseFileTail(srcBytes []byte, cursor int) (newCursor int, err error) {
	defer func() {
		if p := recover(); p != nil {
			newCursor = 0
			err = fmt.Errorf("%v", p)
		}
	}()

	expectedChecksum, err := readNextUint32(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	curChecksum := crc32.Checksum(srcBytes[0:cursor], crc32q)
	if curChecksum != expectedChecksum {
		return 0, ErrChecksumNotMatched
	}
	cursor += Uint32ByteSize

	// 文件结束标记
	gotEndFlag, err := readNextUint16(srcBytes, cursor)
	if err != nil {
		return 0, err
	}
	if gotEndFlag != CompressedFileEndFlag {
		return 0, ErrInvalidEndFlag
	}
	cursor += Uint16ByteSize

	return cursor, nil
}
