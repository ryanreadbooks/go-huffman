package huffman

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompressBytes(t *testing.T) {
	f, err := os.Open("test/test_data.txt")
	require.Nil(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.Nil(t, err)

	denseData, _, err := CompressBytes(data)
	require.Nil(t, err)
	fmt.Printf("data size = %d, denseData size = %d\n", len(data), len(denseData))
}

func TestCompressFile(t *testing.T) {
	CompressFile("test/test_data.txt", "test/test_data.bin")
}