package history_test

import "bytes"

type bytesState []byte

func (b bytesState) Equal(other bytesState) bool { return bytes.Equal(b, other) }

func (b bytesState) Size() int { return len(b) }
