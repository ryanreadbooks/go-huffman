package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryanreadbooks/go-huffman/huffman"
)

// 压缩或解压缩文件
func main() {
	performCompress := flag.Bool("compress", false, "compress given file")
	performDecompress := flag.Bool("decompress", false, "decompress given file")
	inputFile := flag.String("input", "", "input filename")
	outputFile := flag.String("output", "", "output filename")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("please specify the input filename")
		os.Exit(0)
	}
	if *outputFile == "" {
		fmt.Println("please specify the output filename")
		os.Exit(0)
	}
	if *performCompress && *performDecompress {
		fmt.Println("compress flag and decompress can not be both true")
		os.Exit(0)
	}

	if *performCompress {
		fmt.Println("performing compression...")
		err := huffman.CompressFile(*inputFile, *outputFile)
		if err != nil {
			fmt.Printf("compression failed: %v\n", err)
		} else {
			fmt.Println("compression ok")
		}
	}
	if *performDecompress {
		fmt.Println("performing decompression...")
		err := huffman.DecompressFile(*inputFile, *outputFile)
		if err != nil {
			fmt.Printf("decompression failed: %v\n", err)
		} else {
			fmt.Println("decompression ok")
		}
	}
}
