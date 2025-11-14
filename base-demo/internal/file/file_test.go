package file

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestWithFile(t *testing.T) {
	// 测试参数
	fileSize := int64(50 * 1024 * 1024) // 50MB 测试
	fmt.Println("测试文件大小:", fileSize/1024/1024, "MB")

	// 1. 构造一份原始数据（随机填充）
	src := make([]byte, fileSize)
	_, err := rand.Read(src)
	if err != nil {
		t.Fatalf("无法生成随机数据: %v", err)
	}

	// 2. 构造 chunk channel
	chunkChan := make(chan Chunk, 16)

	// 3. 模拟生产者（分片生产数据并送入 channel）
	go func() {
		defer close(chunkChan)

		var offset int64 = 0

		for offset < fileSize {
			// 从内存池中取出 buffer
			bufPtr := chunkPool.Get().(*[]byte)
			buf := *bufPtr

			size := ChunkSize
			if offset+int64(size) > fileSize {
				size = int(fileSize - offset)
			}

			// copy 数据
			copy(buf[:size], src[offset:offset+int64(size)])

			// 发送 chunk
			chunkChan <- Chunk{
				Offset: offset,
				Data:   buf[:size],
			}

			offset += int64(size)
		}
	}()

	// 4. 执行并发写入
	dstFile := "test_output.bin"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	err = WriteFileConcurrently(ctx, dstFile, chunkChan, fileSize)
	if err != nil {
		t.Fatalf("并发写入失败: %v", err)
	}

	// 5. 验证写出的文件是否一致
	dst, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("读取目标文件失败: %v", err)
	}

	if !bytes.Equal(src, dst) {
		t.Fatalf("文件校验失败：内容不一致")
	}

	fmt.Println("🚀 测试通过：文件内容完全一致！")
}

func TestReadFileChunks(t *testing.T) {
	block := 10 * 1024 * 1024 // 10MB

	ReadFileWithMmap("test_output.bin", block, func(chunk []byte) error {
		fmt.Println("read chunk:", len(chunk))
		return nil
	})
}
