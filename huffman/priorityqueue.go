package huffman

import "container/heap"

// 定义一个优先级队列
type huffmanPQ []*HuffmanNode

// 实现heap.Interface接口
func (pq huffmanPQ) Len() int {
	return len(pq)
}

func (pq huffmanPQ) Less(i, j int) bool {
	// 最小堆
	return pq[i].Weight < pq[j].Weight
}

func (pq huffmanPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *huffmanPQ) Push(x interface{}) {
	node := x.(*HuffmanNode)
	*pq = append(*pq, node)
}

func (pq *huffmanPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return node
}

// HuffmanPQ 是对huffmanPQ的封装以方便使用
type HuffmanPQ struct {
	data huffmanPQ
}

// NewHuffmanPQ 创建并返回一个优先队列
func NewHuffmanPQ() *HuffmanPQ {
	pq := &HuffmanPQ{}
	pq.data = make(huffmanPQ, 0)
	heap.Init(&pq.data)

	return pq
}

// NewHuffmanPQFromUInts 从uint64切片创建并返回一个优先队列
func NewHuffmanPQFromUInts(ints []uint64) *HuffmanPQ {
	pq := &HuffmanPQ{}
	pq.data = make(huffmanPQ, 0, len(ints))

	for i := 0; i < len(ints); i++ {
		pq.data = append(pq.data, &HuffmanNode{Weight: ints[i]})
	}

	heap.Init(&pq.data)

	return pq
}

// ?? 在初始化之后这个方法是否有影响 ??
func (pq *HuffmanPQ) UpdateOrder() {
	heap.Init(&pq.data)
}

func (pq *HuffmanPQ) Push(node *HuffmanNode) {
	heap.Push(&pq.data, node)
}

func (pq *HuffmanPQ) Pop() *HuffmanNode {
	return heap.Pop(&pq.data).(*HuffmanNode)
}

func (pq *HuffmanPQ) Size() int {
	return pq.data.Len()
}

func (pq *HuffmanPQ) Peek() *HuffmanNode {
	return pq.data[0]
}
