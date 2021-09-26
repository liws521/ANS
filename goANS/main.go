package main

import (
	"fmt"
	"goANS/utils/histogram"
	"goANS/utils/rans"
)

func main() {
	var tablelog uint = 3
	seq := "abbcbcdcc"
	ct := histogram.Count(seq)
	//for k, v := range ct {
	//	fmt.Println(k, "->", v)
	//}
	nt := histogram.NormalizedCount(ct, tablelog)
	for k, v := range nt {
		fmt.Println(k, "->", v)
	}
	cumul, symbolTable := histogram.Cumulative(nt, tablelog)
	for k, v := range cumul {
		fmt.Println(k, "->", v)
	}
	for _, v := range symbolTable {
		fmt.Println(v)
	}
	m := rans.NewMessage()
	encoder, decoder := rans.RAns(cumul, nt, tablelog, symbolTable)
	for i := 0; i < len(seq); i++ {
		m = encoder(m, seq[i])
	}

	decodeSeq := make([]byte, 0, len(seq))
	for i := 0; i < len(seq); i++ {
		var sym byte
		m, sym = decoder(m)
		decodeSeq = append(decodeSeq, sym)
	}
	for _, v := range decodeSeq {
		fmt.Println(v)
	}
}
