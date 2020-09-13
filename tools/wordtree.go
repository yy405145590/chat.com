package tools

type wordNode struct {
	children map[byte]*wordNode
	isEnd    bool
}

func newWordNode() *wordNode {
	return &wordNode{children: make(map[byte]*wordNode), isEnd: false}
}

type WordTree struct {
	root *wordNode
}

func NewRoot() *WordTree {
	return &WordTree{root: newWordNode()}
}

func (tree *WordTree) Insert(word string) {
	node := tree.root
	for i := 0; i < len(word); i++ {
		_, ok := node.children[word[i]]
		if !ok {
			node.children[word[i]] = newWordNode()
		}
		node = node.children[word[i]]
	}
	node.isEnd = true
}

func (tree *WordTree) Search(word string) bool {
	node := tree.root
	for i := 0; i < len(word); i++ {
		_, ok := node.children[word[i]]
		if !ok {
			return false
		}
		node = node.children[word[i]]
	}
	return node.isEnd
}
