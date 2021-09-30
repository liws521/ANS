package model

import (
	"fmt"
	"goANS/utils/myerr"
	"goANS/utils/myio"
)

// 关于各个count表中的类型设计成int还是uint的问题,
// 因为fse中还有normalizedCount表中出现-1的情况, 所以做成int吧
// 有需要与uint进行运算的位置就做一下强制类型转换

const (
	MAX_SYMBOL = 255
)

const (
	RANS = iota
	TANS
)

// 概率模型结构体
// 给定一个src待压缩字符序列, 生成一个概率模型
type Model struct {
	ct        [MAX_SYMBOL + 1]int // Count Table
	Nt        [MAX_SYMBOL + 1]int // Normalized Table
	Cumul     [MAX_SYMBOL + 1]int // Cumulative Table
	Ts        []byte              // table of symbol
	mark      int                 // 标记是用于哪种压缩算法的概率模型, 对Ts的分布有影响
	maxSymbol int
	Tablelog  int
	Total     int
}

func NewModel(src []byte, mark int) *Model {
	// 如果src是nil或len==0, 无需生成概率模型
	if src == nil || len(src) == 0 {
		return nil
	}
	m := &Model{
		mark:     mark,
		Tablelog: 5,
		Total:    len(src),
	}
	// 首先对src进行符号频次统计, 填写ct表
	// Count的功能比较通用, 所以设计成函数就行, 不用设计成绑定model的方法
	Count(src, m.ct[:])
	// 找到出现的最大符号maxSymbol, 将后序操作从MAX_SYMBOL简化到maxSymbol
	m.MaxSymbol()
	// 接下来的任务是判断是否src中只有一个符号
	// 可以单独写一个函数判断, 也可以在normalized时判断 if(ct[i] == Total)
	// 这里采用第二种方式
	// 还有一个任务是根据srcSize与maxSymbol选择合适的Tablelog, 这里先用5

	// 接下来进行normalizedCount
	tmp := m.NormalizedCount()
	if tmp == -1 {
		return nil
	}
	// fill cumulative table
	m.Cumulative()
	// 根据ANS的类型的不同进行不同的Spread
	m.Spread()
	// 返回生成的概率模型
	return m
}

// Count the number of each symbol
func Count(src []byte, ct []int) {
	for _, v := range src {
		ct[int(v)]++
	}
	return
}

// MaxSymbol
func (m *Model) MaxSymbol() {
	for i := MAX_SYMBOL; i >= 0; i-- {
		if m.ct[i] > 0 {
			m.maxSymbol = i
		}
	}
	// ct表中全为0的case应该提前判断拒绝生成概率模型, 所以不会执行到这里
}

func (m *Model) NormalizedCount() int {
	scale, step := 29-m.Tablelog, (1<<30)/m.Total
	stillToDistribute := 1 << m.Tablelog
	largestS, largestP := 0, 0
	lowThreshold := m.Total >> m.Tablelog
	for i, v := range m.ct {
		if v > largestP {
			largestS = i
			largestP = v
		}
		if v == m.Total {
			fmt.Println("RLE")
			return -1
		}
		if v == 0 {
			continue
		}
		if v < lowThreshold {
			m.Nt[i] = 1
			stillToDistribute--
		} else {
			// 多乘了一个2, 为了判断四舍五入
			freq := (v * step) >> scale
			remainder := freq % 2
			freq /= 2
			if freq < 8 {
				freq += remainder
			}
			m.Nt[i] = freq
			stillToDistribute -= freq
		}
	}
	m.Nt[largestS] += stillToDistribute
	return 0
}

// Cumulative
func (m *Model) Cumulative() {
	start := 0
	for i, v := range m.Nt {
		m.Cumul[i] = start
		start += v
	}
	return
}

// Spread
func (m *Model) Spread() {
	switch m.mark {
	case RANS:
		m.Ts = make([]byte, 1<<m.Tablelog)
		pos := 0
		for i, v := range m.Nt {
			for j := 0; j < v; j++ {
				m.Ts[pos] = byte(i)
				pos++
			}
		}
		myerr.Assert(pos == 1<<m.Tablelog, "rANS Spread failed")
	default:
	}
}

// WrToByte
// 需要把概率模型信息写入到压缩后的文件中, 解压时读取概率模型构造解码器
// 这里把信息写到一个[]byte中
// 为了实现简单, 先不考虑压缩率, 压一个字节的tablelog, 256个字节的normalizedTable
// 还需要把四个字节的原文大小, 也就是Total写入
func (m *Model) WrToByte() []byte {
	bs := make([]byte, MAX_SYMBOL+6)
	totals := myio.UintToByteSlice(uint(m.Total))
	for i := 0; i < 4; i++ {
		bs[i] = totals[i]
	}
	id := 4
	bs[id] = byte(m.Tablelog)
	id++
	for _, v := range m.Nt {
		bs[id] = byte(v)
		id++
	}
	return bs
}

// RdFromByte
// WrToByte的逆操作, 从[]byte中读取信息, 构造一个概率模型并返回
func RdFromByte(src []byte) (*Model, int) {
	m := &Model{
		mark: RANS,
	}
	m.Total = int(myio.ByteSliceToUint(src[:4]))
	id := 4
	m.Tablelog = int(src[id])
	id++
	for i := 0; i < MAX_SYMBOL+1; i++ {
		m.Nt[i] = int(src[id])
		id++
	}
	m.Cumulative()
	m.Spread()
	return m, id
}
