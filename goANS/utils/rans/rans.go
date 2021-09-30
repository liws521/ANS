package rans

import (
	"fmt"
	"goANS/utils/datatype"
	"goANS/utils/model"
	"goANS/utils/myio"
)

// 编码器和解码器的实现可以做成闭包的实现对概率模型进行封装

var (
	sPrec uint = 64
	tPrec uint = 32
	tMask uint = (1 << tPrec) - 1
	sMin  uint = 1 << (sPrec - tPrec)
	sMax  uint = 1 << sPrec
)

type message struct {
	Val  uint
	Tail *datatype.Stack
}

func NewMessage() *message {
	m := &message{}
	m.Val = sMin
	m.Tail = &datatype.Stack{}
	return m
}

// RAns, 传入一个概率模型, 返回编码器和解码器
// rans的概率分布由 start + prob组成
func RAns(pm *model.Model) (encoder func(m *message, sym byte) *message, decoder func(m *message) (*message, byte)) {
	encoder = func(m *message, sym byte) *message {
		start, prob := uint(pm.Cumul[sym]), uint(pm.Nt[sym])
		for m.Val >= prob<<(sPrec-uint(pm.Tablelog)) {
			m.Tail.Push(m.Val & tMask)
			m.Val >>= tPrec
		}
		m.Val = (m.Val / prob << uint(pm.Tablelog)) + m.Val%prob + start
		return m
	}
	decoder = func(m *message) (*message, byte) {
		var sBar uint = m.Val & ((1 << uint(pm.Tablelog)) - 1)
		sym := pm.Ts[sBar]
		var start, prob uint = uint(pm.Cumul[sym]), uint(pm.Nt[sym])
		m.Val = prob*(m.Val>>uint(pm.Tablelog)) + sBar - start
		for m.Val < sMin {
			top := m.Tail.Top()
			m.Tail.Pop()
			m.Val = (m.Val << tPrec) + top
		}
		return m, sym
	}
	return encoder, decoder
}

// EncFlush 压缩完所有符号后, 把m.Val当作两个uint32压入栈中, 方便后序将压缩结果写入文件
// 先压高32位, 后压低32位
func (m *message) EncFlush() {
	m.Tail.Push(m.Val >> tPrec)
	m.Tail.Push(m.Val & tMask)
}

func (m *message) FlattenStack() []byte {
	dst := make([]byte, 0)
	for !m.Tail.IsEmpty() {
		top := m.Tail.Top()
		m.Tail.Pop()
		bs := myio.UintToByteSlice(top)
		dst = append(dst, bs[0], bs[1], bs[2], bs[3])
	}
	return dst
}

// DecInit 从src中恢复出message信息
func DecInit(src []byte) *message {
	srcSize := len(src)
	if srcSize == 0 || srcSize % 4 != 0 {
		return nil
	}
	uintsSize := srcSize / 4
	// 首先把[]byte数组每四个一组转成[]uint
	uints := make([]uint, uintsSize)
	for i := 0; i < uintsSize; i++ {
		uints[i] = myio.ByteSliceToUint(src[i * 4: i*4+4])
	}
	m := &message{}
	m.Val = uints[1] << tPrec | uints[0]
	m.Tail = &datatype.Stack{}
	for i := uintsSize - 1; i >= 2; i-- {
		m.Tail.Push(uints[i])
	}
	return m
}

// Compress
func Compress(src []byte) []byte {
	srcSize := len(src)
	// 传入一个src, 根据数据建立概率模型
	pm := model.NewModel(src, model.RANS)
	// 将概率模型的信息写入dst
	dst := pm.WrToByte()
	// 初始化一个message
	m := NewMessage()
	// 将概率模型传入RAns, 获得一个与该模型绑定的编码器
	encoder, _ := RAns(pm)
	// 正向逐字节编码
	for i := 0; i < srcSize; i++ {
		m = encoder(m, src[i])
	}
	// 将m.Val的信息flush到栈中
	m.EncFlush()
	// 把数据从栈中flatten到[]byte中
	return append(dst, m.FlattenStack()...)
}

// Decompress
// 写解码的时候发现, 压缩的时候需要把概率模型压缩进去, 把tablelog和nt压进去就行了
func Decompress(src []byte) []byte {
	// 根据约定好的协议, 解析出tablelog和normalizedTable, 构造概率模型
	pm, n := model.RdFromByte(src[:])
	// 从src剩下的流中恢复出message
	m := DecInit(src[n:])
	_, decoder := RAns(pm)
	dst := make([]byte, pm.Total)
	var sym byte
	for i := 0; i < pm.Total; i++ {
		m, sym = decoder(m)
		dst[pm.Total - 1 - i] = sym
	}

	//for i, j := 0, len(decodeSeq) - 1; i < j; i, j = i + 1, j - 1 {
	//	decodeSeq[i], decodeSeq[j] = decodeSeq[j], decodeSeq[i]
	//}

	for _, v := range string(dst) {
		fmt.Printf("%c", v)
	}
	return dst
}