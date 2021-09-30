package myio

import (
	"bufio"
	"os"
	"path/filepath"
)

// OpenFile, 传入一个文件路径, 打开文件, 若路径中包含不存在的目录, 则创建
func OpenFile(path string, mode int) (*os.File, error) {
	dirPath := filepath.Dir(path)
	// if dirPath is not exist, make it
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return nil, err
		}
	}
	var file *os.File
	var err error
	if mode == os.O_TRUNC {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	} else {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	}
	if err != nil {
		return nil, err
	}
	return file, err
}

// Read, 封装, 传入一个文件路径, 返回一个[]byte
func Read(path string) ([]byte, error) {
	// 根据路径打开文件
	file, err := OpenFile(path, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 此时文件必定存在了, 因为若不存在也创建了
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, fileInfo.Size())
	reader := bufio.NewReader(file)
	_, err = reader.Read(buf)
	return buf, nil
}

// Write, 传入dst文件路径, 和[]byte, 将其写入文件
// 写的时候如果不想追加, 要清空文件
func Write(path string, dst []byte) (err error) {
	// 根据路径打开文件
	file, err := OpenFile(path, os.O_TRUNC)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write(dst)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}