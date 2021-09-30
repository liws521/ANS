package datatype

import "fmt"

// 存储uint的栈
type Stack struct {
	data []uint
	size int
}

func (p *Stack) IsEmpty() bool {
	return p.size == 0
}

func (p *Stack) Push(v uint) {
	if p.size < len(p.data) {
		p.data[p.size] = v
	} else {
		p.data = append(p.data, v)
	}
	p.size++
}

func (p *Stack) Pop() {
	if p.IsEmpty() {
		return
	} else {
		p.size--
	}
}

func (p *Stack) Top() uint {
	if p.IsEmpty() {
		fmt.Println("Stack is empty, top failed")
		return 0
	} else {
		return p.data[p.size-1]
	}
}
