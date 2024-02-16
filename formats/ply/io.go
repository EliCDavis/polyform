package ply

import "io"

func readByte(in io.Reader) (byte, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(in, buf)
	return buf[0], err
}
