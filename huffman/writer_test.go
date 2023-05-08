package huffman

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitsWriter_Primary(t *testing.T) {
	// code1 := NewHuffmanCodeFromString("1001011")
	// code2 := NewHuffmanCodeFromString("10101111")
	// code3 := NewHuffmanCodeFromString("0001111111")
	// code4 := NewHuffmanCodeFromString("1111100010101")

	s1 := "01010101"
	s2 := "11111111"
	s3 := "00000"
	s4 := "111101"

	code1 := NewHuffmanCodeFromString(s1)
	code2 := NewHuffmanCodeFromString(s2)
	code3 := NewHuffmanCodeFromString(s3)
	code4 := NewHuffmanCodeFromString(s4)

	w := NewBitsWriter()
	w.WriteUint16(code1.Bits(), uint8(code1.BitLen()))
	w.WriteUint16(code2.Bits(), uint8(code2.BitLen()))
	w.WriteUint16(code3.Bits(), uint8(code3.BitLen()))
	w.WriteUint16(code4.Bits(), uint8(code4.BitLen()))

	b := w.Buf()
	// 1001011 10101111 0001111111 1111100010101
	// 1001011 10101111 0001111111 1111100010101
	// 10010111 01011110 00111111 11111100 01010100
	// 0x97		0x5e	 0x3f	  0xfc	   0x54
	s := BytesToString(b, code1.BitLen()+code2.BitLen()+code3.BitLen()+code4.BitLen())
	req := s1 + s2 + s3 + s4
	fmt.Println("got :", s)
	fmt.Println("want:", req)
	// 01010101 11111111 00000111 10100000
	// 55  		ff 		 07		  a0
}

func TestBitsWriter(t *testing.T) {
	tests := []struct {
		Codes []string
	}{
		{Codes: []string{"01010101", "11111111", "00000", "111101"}},
		{Codes: []string{"10101010", "00001111", "1111111111111110", "10101010", "0101010111110000"}},
		{Codes: []string{"1010101010001101", "10101010", "0000000011111111"}},
		{Codes: []string{"1", "0", "0", "1", "0", "1", "0", "1", "010111"}},
		{Codes: []string{"010101", "01010101", "1111110111010101", "0101"}},
		{Codes: []string{"1010101010", "0101010101", "01001100"}},
		{Codes: []string{"0000", "000000001", "111010", "01011010101010"}},
		{Codes: []string{"111111111100", "000", "00000000", "1111111111000000"}},
		{Codes: []string{"1", "0", "1", "0", "0101010101", "0001111"}},
	}

	for _, tc := range tests {
		w := NewBitsWriter()
		expect := strings.Builder{}
		var total int = 0
		for _, s := range tc.Codes {
			expect.WriteString(s)
			huff := NewHuffmanCodeFromString(s)
			bitlen := uint8(huff.BitLen())
			total += int(bitlen)
			w.WriteUint16(huff.Bits(), bitlen)
		}
		b := w.Buf()

		require.EqualValues(t, expect.String(), BytesToString(b, total))
	}
}

func TestBitsWriter_WithError(t *testing.T) {
	w := NewBitsWriter()

	require.NotNil(t, w.WriteUint16(0x9600, 20))
}
