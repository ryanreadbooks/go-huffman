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
		err := w.WriteUint16(code.Bits(), uint8(bitlen))
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
//   - START_FLAG						2 bytes uint16
//   - SRC_FILENAME_LEN					2 bytes uint16
//   - BYTE SIZE BEFORE COMPRESSION		4 byte uint32
//   - BYTE SIZE AFTER COMPRESSION		4 bytes uint32
//   - SRC_FILENAME						n bytes
//
// DATA
//   - HUFFMAN TABLE
//   - COMPRESSED DATA
//   - VALID BIT LEN		8 bytes uint64
//   - COMPRESSED BIT
//
// TAIL
//   - CRC32 CHECKSUM	  	4 bytes uint32
//   - END_FLAG				2 bytes uint16
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

	var alloc uint32 = 26 + compressedSize + uint32(len(encTableSer))
	dstBytes := make([]byte, 0, alloc)
	// 写入文件头
	dstBytes = writeUint16ToBytes(CompressedFileStartFlag, dstBytes) // 文件开始标识
	dstBytes = writeUint16ToBytes(filenameNoDirSize, dstBytes)       // 文件名长度
	dstBytes = writeUint32ToBytes(originalSize, dstBytes)            // 压缩前字节大小
	dstBytes = writeUint32ToBytes(compressedSize, dstBytes)          // 压缩后字节大小
	dstBytes = append(dstBytes, []byte(filenameNoDir)...)            // 源文件名

	// 写入数据区
	dstBytes = append(dstBytes, encTableSer...)     // Huffman码表
	dstBytes = writeUint64ToBytes(bitLen, dstBytes) // bitlen
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
	return nil
}
