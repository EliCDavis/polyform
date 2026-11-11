package ply

import (
	"bytes"
	"io"
	"runtime"
	"strconv"

	"github.com/EliCDavis/polyform/modeling"
)

func writeVertices(out io.Writer, mesh modeling.Mesh, writers []PropertyWriter, format Format, count int) error {
	if count == 0 || len(writers) == 0 {
		return nil
	}

	built := make([]builtPropertyWriter, len(writers))
	rowSize := 0
	for i, w := range writers {
		built[i] = w.build(mesh, format)
		for _, p := range w.Properties(mesh) {
			rowSize += p.(ScalarProperty).Size()
		}
	}

	var separators []byte
	if format == ASCII {
		separators = bytes.Repeat([]byte{' '}, len(built))
		separators[len(built)-1] = '\n'
	}

	rowsPerChunk := max(1, binaryVertexChunkBytes/rowSize)
	workers := runtime.NumCPU()

	free := make(chan []byte, workers+2)
	for range cap(free) {
		free <- make([]byte, 0, rowsPerChunk*rowSize)
	}

	var catcher panicCatcher
	results := make(chan chan []byte, workers)
	go func() {
		defer close(results)
		for start := 0; start < count && !catcher.failed(); start += rowsPerChunk {
			buf := (<-free)[:0]
			end := min(start+rowsPerChunk, count)
			done := make(chan []byte, 1)
			results <- done
			go func() {
				catcher.run(func() {
					for i := start; i < end; i++ {
						for j, w := range built {
							buf = w.Append(buf, i)
							if separators != nil {
								buf = append(buf, separators[j])
							}
						}
					}
				})
				done <- buf
			}()
		}
	}()

	var err error
	for done := range results {
		buf := <-done
		if err == nil && !catcher.failed() {
			_, err = out.Write(buf)
		}
		free <- buf
	}
	catcher.rethrow()
	return err
}

func appendAsciiInts(dst []byte, values ...float64) []byte {
	for i, v := range values {
		if i > 0 {
			dst = append(dst, ' ')
		}
		dst = strconv.AppendInt(dst, int64(v), 10)
	}
	return dst
}

func appendAsciiFloats(dst []byte, values ...float64) []byte {
	for i, v := range values {
		if i > 0 {
			dst = append(dst, ' ')
		}
		dst = strconv.AppendFloat(dst, v, 'f', -1, 64)
	}
	return dst
}
