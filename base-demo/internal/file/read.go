package file

import (
	"os"
	"syscall"
)

// ReadFileWithMmap mmap 分片读取实现（无额外内存 copy）
func ReadFileWithMmap(path string, blockSize int, handler func(chunk []byte) error) error {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件大小
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fi.Size()

	if fileSize == 0 {
		return nil
	}

	data, err := syscall.Mmap(
		int(file.Fd()),
		0,
		int(fileSize),
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return err
	}
	defer syscall.Munmap(data)

	for offset := 0; offset < len(data); offset += blockSize {
		end := min(offset+blockSize, len(data))

		chunk := data[offset:end] // 🔥 直接引用 mmap 区域，不复制

		if err := handler(chunk); err != nil {
			return err
		}
	}
	return nil
}
