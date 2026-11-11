package ply

import (
	"io"
	"runtime"
	"sync"
)

const binaryVertexChunkBytes = 1 << 20

func readBinaryVertices(in io.Reader, readers []binaryPropertyReader, count, rowSize int) error {
	if count == 0 || rowSize == 0 {
		return nil
	}

	type chunk struct {
		start int
		buf   []byte
	}

	rowsPerChunk := max(1, binaryVertexChunkBytes/rowSize)
	workers := runtime.NumCPU()

	free := make(chan []byte, workers+1)
	for range cap(free) {
		free <- make([]byte, rowsPerChunk*rowSize)
	}

	var catcher panicCatcher
	jobs := make(chan chunk)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				catcher.run(func() {
					for r := 0; r < len(c.buf)/rowSize; r++ {
						row := c.buf[r*rowSize : (r+1)*rowSize]
						for _, reader := range readers {
							reader.Read(row, int64(c.start+r))
						}
					}
				})
				free <- c.buf
			}
		}()
	}

	var err error
	for start := 0; start < count && !catcher.failed(); start += rowsPerChunk {
		buf := <-free
		rows := min(rowsPerChunk, count-start)
		buf = buf[:rows*rowSize]
		if _, err = io.ReadFull(in, buf); err != nil {
			break
		}
		jobs <- chunk{start: start, buf: buf}
	}
	close(jobs)
	wg.Wait()
	catcher.rethrow()
	return err
}
