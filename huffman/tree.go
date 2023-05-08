package huffman

// HuffmanNode 表示一个Huffman树的节点
type HuffmanNode struct {
	Weight uint64
	Parent *HuffmanNode
	Left   *HuffmanNode
	Right  *HuffmanNode
	Byte   byte
	Code   *HuffmanCode
}

// IsLeaf 判断当前节点是否为叶子节点
func (nd *HuffmanNode) IsLeaf() bool {
	return nd.Left == nil && nd.Right == nil
}

// IsLeft 判断当前节点是否为左子树
func (nd *HuffmanNode) IsLeft() bool {
	if nd.Parent != nil {
		return nd.Parent.Left == nd
	}
	return false
}

// IsRight 判断当前节点是否为右子树
func (nd *HuffmanNode) IsRight() bool {
	if nd.Parent != nil {
		return nd.Parent.Right == nd
	}
	return false
}

// setCode 设置这个节点的Huffman编码
// 设置思路：从叶子节点不断往上，追加比特位，最后将所有比特位逆序
func (nd *HuffmanNode) setCode() {
	cur := nd
	huffmanBits := &HuffmanCode{}
	for cur != nil {
		if cur.IsLeft() {
			huffmanBits.AppendZero()
		} else if cur.IsRight() {
			huffmanBits.AppendOne()
		}
		cur = cur.Parent
	}
	nd.Code = huffmanBits.ReverseNew()
}

// WeightLength 计算带权路径长度
// 这个方法仅在nd.Code被设置了后才能得到有效值
func (nd *HuffmanNode) WeightLength() int {
	return int(nd.Weight) * nd.Code.BitLen()
}

// HuffmanTree 表示了一棵Huffman树
type HuffmanTree struct {
	Freq   Frequencies
	Root   *HuffmanNode
	Leaves []*HuffmanNode
}

// NewHuffmanTree 根据指定频率构建一棵新的Huffman树
func NewHuffmanTree(freq Frequencies) *HuffmanTree {
	root, leaves := ConstructHuffmanTree(freq)
	tree := &HuffmanTree{
		Freq:   freq,
		Root:   root,
		Leaves: leaves,
	}

	return tree
}

// ConstructHuffmanTree 根据频率创建一棵Huffman树
// 返回Huffman树的根节点和所有叶子节点
func ConstructHuffmanTree(freq Frequencies) (*HuffmanNode, []*HuffmanNode) {
	// 单独处理只有一个数据的情况
	if len(freq) == 1 {
		var k byte
		var v uint64
		for k, v = range freq {
		}
		root := &HuffmanNode{}
		left := &HuffmanNode{Parent: root, Weight: v, Byte: k}
		root.Left = left

		left.Code = NewHuffmanCodeFromString("0")
		return root, []*HuffmanNode{left}
	}

	// 1. 构建优先级队列
	pq := NewHuffmanPQ()

	// 所有叶子节点
	var leaves []*HuffmanNode = make([]*HuffmanNode, 0, len(freq))
	for k, v := range freq {
		node := &HuffmanNode{Weight: v, Byte: k}
		leaves = append(leaves, node)
		pq.Push(node)
	}

	// 2. 开始构建Huffman树
	for pq.Size() > 1 {
		// pop出weight最小的两个节点
		nodeA := pq.Pop()
		nodeB := pq.Pop()
		if (nodeA.Weight == nodeB.Weight) && (nodeA.Byte > nodeB.Byte) {
			nodeA, nodeB = nodeB, nodeA
		}
		// 新增一个根节点合并这个节点
		nodeRoot := &HuffmanNode{
			Left:   nodeA,
			Right:  nodeB,
			Weight: nodeA.Weight + nodeB.Weight,
		}
		nodeA.Parent = nodeRoot
		nodeB.Parent = nodeRoot
		// 将新增的节点重新插回优先级队列中
		pq.Push(nodeRoot)
	}

	// 给叶子节点编码
	for _, leaf := range leaves {
		leaf.setCode()
	}

	return pq.Peek(), leaves
}
