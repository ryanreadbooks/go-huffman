package huffman

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHuffmanTable_SerializeAndDeserialize(t *testing.T) {
	f, err := os.Open("test/test_data.txt")
	require.Nil(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.Nil(t, err)

	// 获得HuffmanTable
	freq := CountFrequencies(data)
	tree := NewHuffmanTree(freq)
	table := NewHuffmanEncTable(tree)
	// 序列化
	ser, err := table.Serialize()
	require.Nil(t, err)

	tableF, err := os.Create("test/table.bin")
	require.Nil(t, err)
	defer os.Remove("test/table.bin")
	defer tableF.Close()

	n, err := tableF.Write(ser)
	require.Nil(t, err)
	require.EqualValues(t, n, len(ser))

	// 反序列化
	deTable, err := DeserializeHuffmanEncTable(ser)
	require.Nil(t, err)
	require.Equal(t, deTable.ItemNum(), table.ItemNum())
	require.True(t, deTable.Equals(table))
	require.True(t, table.Equals(deTable))
}
