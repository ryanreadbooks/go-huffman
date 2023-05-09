# go-huffman

该仓库使用Golang直观地实现Huffman编码，并且实现基于Huffman编码的压缩和解压缩。

## 使用方法

### Build

```bash
go build -o example
```

### 压缩

```bash
./example -compress -input 需要压缩文件名 -output 目标文件名
```

### 解压

```bash
./example -decompress -input 需要解压文件名 -output 目标文件名
```

## 实现细节

### Huffman编码

#### 编码过程

1. 以字节为基本单位，计算频数；

2. 使用优先队列构建Huffman二叉树；

3. 对于每个叶子节点，往上遍历到根节点，记录编码比特（左0右1），最后将比特位逆序得到叶子节点的编码结果。

#### 编码存储

​	本仓库先支持最大长度为24bit的Huffman编码（可以存在一个32bit整数中）。具体为：高8位存Huffman编码的比特长度；低24位存Huffman编码本身，Huffman编码本身的最高位放在低24位的最高位。

<img src="/media/ryan/Documents/Codes/GoCodes/go-huffman/docs/image/huffmancode.png" alt="huffman-code-format" style="float:left;zoom:30%;" />

### 文件格式

#### 压缩文件格式

压缩文件格式如下，包含文件头HEADER，数据区DATA和文件尾TAIL。

文件头HEADER包含开始标记等信息；数据区DATA包含Huffman码表和压缩后的数据；文件尾TAIL包含校验和和结束标记。

```
压缩文件格式如下：（大端序）
HEADER
	- START_FLAG						2 bytes (uint16)
	- SRC_FILENAME_LEN					2 bytes (uint16)
	- BYTE SIZE BEFORE COMPRESSION		4 bytes (uint32)
	- BYTE SIZE AFTER COMPRESSION		4 bytes (uint32)
	- SRC_FILENAME						n bytes

DATA
	- HUFFMAN TABLE
		-- HUFFMAN TABLE SIZE 		4 bytes (uint32)
		-- HUFFMAN TABLE DATA
	- COMPRESSED DATA
		-- VALID BIT LEN			4 bytes (uint32) + 1 bytes = 5 bytes
		-- COMPRESSED BIT

TAIL
	- CRC32 CHECKSUM	  	4 bytes (uint32)
	- END_FLAG				2 bytes (uint16)
```

#### Huffman码表存储格式

Huffman码表在文件中的存储格式如下，

```
序列化格式如下（大端序：高位放在低地址，低位放在高地址）
START_FLAG					4 bytes
NUMBER OF TABLE ITEMS		4 bytes (uint32)
TABLE_ITEM_1(BYTE+CODE)		1+4=5 bytes
TABLE_ITEM_2(BYTE+CODE)		1+4=5 bytes
...
TABLE_ITEM_N(BYTE+CODE)		1+4=5 bytes
CRC32						4 bytes
END_FLAG					4 bytes
```

## 已知问题

1. Huffman编码最大长度为24bit，如果构建出来的Huffman树高度大于等于25层，则编码错误，从而导致解压缩后的文件与源文件不一致。如果极端情况下二叉树的每一层仅有一个叶子节点，那么Huffman树就会很高（编码字节的话最大257（256+1）层？）。要解决这个问题，则需要用更大的比特位存Huffman编码。
2. Huffman码表存储格式中，为了方便解码，使用的是固定大小的表项（5 bytes）。此处其实也可以仅存储有效的Huffman编码比特位，可以稍微节省一点存储空间，但是这样解码操作就会复杂一点。
3. 数据编解码没有考虑对齐。