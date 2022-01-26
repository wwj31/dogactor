package compress

import (
	"bytes"
	zip "github.com/klauspost/compress/zlib"
	"github.com/klauspost/pgzip"
	"io"
	"sync"
)

var byteBufferPool = sync.Pool{New: func() interface{}{return new(bytes.Buffer)}}
var pgzipWritePool = sync.Pool{New: func() interface{}{return new(pgzip.Writer)}}
var zipWritePool = sync.Pool{New: func() interface{}{return new(zip.Writer)}}

func Compress(data []byte) ([]byte, error) {
	buffer := byteBufferPool.Get().(*bytes.Buffer)
	write := zipWritePool.Get().(*zip.Writer)
	write.Reset(buffer)
	defer func() {
		buffer.Reset()
		byteBufferPool.Put(buffer)
		zipWritePool.Put(write)
	}()

	writer := zip.NewWriter(buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Compresspg 并发压缩，性能不稳定
func Compresspg(data []byte,blockSize,blocks int) ([]byte, error) {
	buffer := byteBufferPool.Get().(*bytes.Buffer)
	write := pgzipWritePool.Get().(*pgzip.Writer)
	write.Reset(buffer)
	defer func() {
		buffer.Reset()
		byteBufferPool.Put(buffer)
		pgzipWritePool.Put(write)
	}()

	writer := pgzip.NewWriter(buffer)
	if err := writer.SetConcurrency(blockSize,blocks);err != nil {
		return nil,err
	}
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}


func Uncompress(data []byte) ([]byte, error) {
	gzipReader, err := zip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	buffer := byteBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		byteBufferPool.Put(buffer)
		_ = gzipReader.Close()
	}()

	_, err = io.Copy(buffer, gzipReader)
	if err != nil {
		return nil, err
	}

	size := len(buffer.Bytes())
	output := make([]byte, size, size)
	copy(output, buffer.Bytes())
	return output, nil
}