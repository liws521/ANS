package rans

import (
	"fmt"
	"goANS/utils/myio"
	"testing"
)

func TestCompress(t *testing.T) {
	// srcFile -> src []byte
	srcFilePath := ".\\tmp\\a.txt"
	src, err := myio.Read(srcFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// src -> Compress -> dst
	dst := Compress(src)

	// dst []byte -> dstFile
	dstFilePath := ".\\tmp\\a.txt.rans"
	myio.Write(dstFilePath, dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

//func TestDecompress(t *testing.T) {
//	// compressedFile -> src []byte
//	srcFilePath := ".\\tmp\\a.tat.rans"
//	src, err := myio.Read(srcFilePath)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	// src -> dst
//	dst := Decompress(src)
//
//	// dst []byte -> file
//	dstFilePath := ".\\tmp\\a.txt.new"
//	myio.Write(dstFilePath, dst)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	return
//}