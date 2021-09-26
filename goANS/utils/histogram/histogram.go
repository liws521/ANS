package histogram
// 关于各个count表中的类型设计成int还是uint的问题,
// 因为fse中还有normalizedCount表中出现-1的情况, 所以做成int吧
// 有需要与uint进行运算的位置就做一下强制类型转换

import "sort"

// Count the frequency of each symbol
func Count(seq string) (ct map[byte]uint) {
	ct = make(map[byte]uint)
	for i := 0; i < len(seq); i++ {
		cur, ok := ct[seq[i]]
		if ok {
			ct[seq[i]] = cur + 1
		} else {
			ct[seq[i]] = 1
		}
	}
	return
}

func NormalizedCount(ct map[byte]uint, tablelog uint) (nt map[byte]uint) {
	var remaining uint = 1 << tablelog
	var total uint = 0
	var largestS byte
	var largestP uint = 0
	for _, v := range ct {
		total += v
	}
	nt = make(map[byte]uint)
	for k, v := range ct {
		freq := (v << tablelog) / total
		if freq < 1 {
			freq = 1
		}
		if freq > largestP {
			largestS = k
			largestP = freq
		}
		nt[k] = freq
		remaining -= freq
	}
	nt[largestS] += remaining
	return
}

// Cumulative
// 因为Golang的map遍历是随机遍历的, 这里要么处理一下让它按序输出, 要么就不管字典序了, 反正也没什么影响
// 至于怎么让它按序遍历呢, 放入slice, 排序, 真的粗暴
func Cumulative(nt map[byte]uint, tablelog uint) (cumul map[byte]uint, symbolTable []byte) {
	cumul = make(map[byte]uint)
	symbolTable = make([]byte, 1 << tablelog)
	bs := make([]byte, 0, len(nt))
	for k := range nt {
		bs = append(bs, k)
	}
	sort.Slice(bs, func(i, j int) bool { return bs[i] < bs[j] })
	start := 0
	for _, v := range bs {
		for i := 0; i < int(nt[v]); i++ {
			symbolTable[start + i] = v
		}
		cumul[v] = start
		start += nt[v]
	}
	return
}
