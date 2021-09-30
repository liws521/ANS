package main

import (
	"fmt"
	"goANS/utils/myio"
	"goANS/utils/rans"
)

func com() {
	// srcFile -> src []byte
	srcFilePath := ".\\utils\\rans\\tmp\\a.txt"
	src, err := myio.Read(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}


	// src -> Compress -> dst
	dst := rans.Compress(src)
	fmt.Println(dst)

	// dst []byte -> dstFile
	dstFilePath := ".\\utils\\rans\\tmp\\a.txt.rans"
	myio.Write(dstFilePath, dst)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func dec() {
	// compressedFile -> src []byte
	srcFilePath := ".\\utils\\rans\\tmp\\a.txt.rans"
	src, err := myio.Read(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// src -> dst
	dst := rans.Decompress(src)

	// dst []byte -> file
	dstFilePath := ".\\utils\\rans\\tmp\\a.txt.new"
	myio.Write(dstFilePath, dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func main() {
	com()
	dec()
}
