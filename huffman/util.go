package huffman

import "strings"

// CountFrequencies 统计输入字节切片中每个字节的出现频数
func CountFrequencies(data []byte) Frequencies {
	frequency := make(Frequencies)
	for _, d := range data {
		frequency[d] += 1
	}

	return frequency
}

// BytesToString 将data中的一共bitsize位比特转化成01字符串
func BytesToString(data []byte, bitsize int) string {
	builder := strings.Builder{}
	builder.Grow(bitsize)
	var scannedLen int = 0
	var mask byte = 1 << 7

	for i := 0; i < len(data); i++ {
		var curSlot int = 0
		bits := data[i]
		for curSlot < 8 && scannedLen < bitsize {
			if (bits & mask) == 0 {
				builder.WriteByte('0')
			} else {
				builder.WriteByte('1')
			}
			bits <<= 1
			curSlot++
			scannedLen++
		}
	}

	return builder.String()
}

// writeUint32ToBytes s将4个字节写入buf中
func writeUint32ToBytes(in uint32, buf []byte) []byte {
	// 大端序
	buf = append(buf, byte((in&0xFF000000)>>24))
	buf = append(buf, byte((in&0x00FF0000)>>16))
	buf = append(buf, byte((in&0x0000FF00)>>8))
	buf = append(buf, byte(in&0x000000FF))

	return buf
}

// writeUint16ToBytes s将2个字节写入buf中
func writeUint16ToBytes(in uint16, buf []byte) []byte {
	// 大端序
	buf = append(buf, byte((in&0xFF00)>>8))
	buf = append(buf, byte((in & 0x00FF)))

	return buf
}

// writeUint64ToBytes s将8个字节写入buf中
func writeUint64ToBytes(in uint64, buf []byte) []byte {
	// 大端序
	buf = append(buf, byte((in&0xFF00000000000000)>>56))
	buf = append(buf, byte((in&0x00FF000000000000)>>48))
	buf = append(buf, byte((in&0x0000FF0000000000)>>40))
	buf = append(buf, byte((in&0x000000FF00000000)>>32))
	buf = append(buf, byte((in&0x00000000FF000000)>>24))
	buf = append(buf, byte((in&0x0000000000FF0000)>>16))
	buf = append(buf, byte((in&0x000000000000FF00)>>8))
	buf = append(buf, byte(in&0x00000000000000FF))

	return buf
}

