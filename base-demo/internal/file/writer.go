package file

import (
	"context"
	"errors"
	"os"
	"sync"
)

const (
	Workers   = 4
	ChunkSize = 10 * 1024 * 1024 // 10MB
)

type Chunk struct {
	Offset int64
	Data   []byte
}

var chunkPool = sync.Pool{
	New: func() any {
		buf := make([]byte, ChunkSize)
		return &buf
	},
}

func WriteFileConcurrently(
	ctx context.Context,
	filePath string,
	chunks <-chan Chunk,
	fileSize int64,
) error {

	// 打开文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 扩容文件
	if err := file.Truncate(fileSize); err != nil {
		return err
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	// 启动 Workers
	for range Workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					errs = append(errs, ctx.Err())
					mu.Unlock()
					return

				case chunk, ok := <-chunks:
					if !ok {
						return
					}

					// 并发写文件（WriteAt 是线程安全的）
					if _, err := file.WriteAt(chunk.Data, chunk.Offset); err != nil {
						mu.Lock()
						errs = append(errs, err)
						mu.Unlock()
					}

					// 将 chunk.Data 放回内存池复用
					if cap(chunk.Data) == ChunkSize { // 防止外部传入的切片
						bufPtr := (*[]byte)(nil)
						buf := chunk.Data[:ChunkSize] // full slice
						bufPtr = &buf
						chunkPool.Put(bufPtr)
					}
				}
			}
		}()
	}

	wg.Wait()
	return errors.Join(errs...)
}
