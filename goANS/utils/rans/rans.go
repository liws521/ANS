package rans

import "fmt"

// 编码器和解码器的实现可以做成闭包的实现对概率模型进行封装

var (
	sPrec uint = 64
	tPrec uint = 32
	tMask uint = (1 << tPrec) - 1
	sMin uint = 1 << (sPrec - tPrec)
	sMax uint = 1 << sPrec
)

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
		return p.data[p.size - 1]
	}
}

type message struct {
	Val uint
	Tail *Stack
}

func NewMessage() *message {
	m := &message{}
	m.Val = sMin
	m.Tail = &Stack{}
	return m
}

// RAns, 传入一个概率模型, 返回编码器和解码器
// rans的概率分布由 start + prob组成
func RAns(cumul, nt map[byte]uint, tablelog uint, symbolTable []byte) (encoder func(m *message, sym byte) (*message), decoder func(m *message) (*message, byte)) {
	encoder = func(m *message, sym byte) (*message) {
		start, prob := uint(cumul[sym]), uint(nt[sym])
		for m.Val >= prob << (sPrec - tablelog) {
			m.Tail.Push(m.Val & tMask)
			m.Val >>= tPrec
		}
		m.Val = (m.Val / prob << tablelog) + m.Val % prob + start
		return m
	}
	decoder = func(m *message) (*message, byte) {
		sBar := m.Val & ((1 << tablelog) - 1)
		sym := symbolTable[sBar]
		start := cumul[sym]
		prob := uint(nt[sym])
		m.Val = prob * (m.Val >> uint(tablelog)) + sBar - start
		for m.Val < sMin {
			top := m.Tail.Top()
			m.Tail.Pop()
			m.Val = (m.Val << tPrec) + top
		}
		return m, sym
	}
	return encoder, decoder
}