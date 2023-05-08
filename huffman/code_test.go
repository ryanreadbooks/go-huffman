package huffman

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func randInt32(min, max int32) int32 {
	return rand.Int31n(max-min) + min
}

// randomOneZeroString 随机生成仅包含0和1的字符串
func randomOneZeroString() string {
	rand.Seed(time.Now().Unix())

	size := randInt32(1, 15)
	var builder strings.Builder
	builder.Grow(int(size))

	for i := 0; i < int(size); i++ {
		if rand.Intn(2) == 0 {
			builder.WriteByte('0')
		} else {
			builder.WriteByte('1')
		}
	}

	return builder.String()
}

func TestHuffmanCode_Primary(t *testing.T) {
	var s = "011010100101001"
	require.EqualValues(t, s, NewHuffmanCodeFromString(s).String())
}

func TestHuffmanCode(t *testing.T) {
	var testCases []string
	for i := 0; i < int(randInt32(10, 100)); i++ {
		testCases = append(testCases, randomOneZeroString())
	}
	for _, tc := range testCases {
		require.EqualValues(t, tc, NewHuffmanCodeFromString(tc).String())
	}

	// 测试非法字符串
	require.Panics(t, func() { NewHuffmanCodeFromString("012301010") })
}

func reverseString(s string) string {
	b := strings.Builder{}
	b.Grow(len(s))
	for i := len(s) - 1; i >= 0; i-- {
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestHuffmanCode_ReversePrimary(t *testing.T) {
	ss := []string{
		"100101001110110",
		"10010",
		"00101101",
		"100100010111",
	}
	for _, s := range ss {
		require.EqualValues(t, reverseString(s), NewHuffmanCodeFromString(s).ReverseNew().String())
	}
}

func TestHuffmanCode_Reverse(t *testing.T) {
	var testCases []string
	for i := 0; i < int(randInt32(1, 100)); i++ {
		testCases = append(testCases, randomOneZeroString())
	}
	for _, tc := range testCases {
		require.EqualValues(t, reverseString(tc), NewHuffmanCodeFromString(tc).ReverseNew().String())
	}

	h := &HuffmanCode{}
	require.EqualValues(t, h.BitLen(), 0)
	require.EqualValues(t, h.BitLen(), h.ReverseNew().BitLen())
}

func TestHuffmanClone(t *testing.T) {
	code := NewHuffmanCodeFromString("10101010101111")
	clone := code.Clone()
	require.EqualValues(t, code.Bits(), clone.Bits())
}
