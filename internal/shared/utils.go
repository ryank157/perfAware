package shared

import (
	"fmt"
	"os"

	"github.com/ryank157/perfAware/internal/timing"
)

type Buffer struct {
	Data  []byte
	Count int64
}

func AllocateBuffer(size int64) Buffer {
	return Buffer{Data: make([]byte, size), Count: size}
}

// struct and func from earlier replies
func FreeBuffer(buf *Buffer) {
	buf.Data = nil // Release the memory held by the buffer's slice
	buf.Count = 0  // Reset the count  VERY important
}

// struct and func from earlier replies
func ReadEntireFile(fileName string) (Buffer, error) {
	defer timing.TimeFunction()()
	data, err := os.ReadFile(fileName)
	if err != nil {
		return Buffer{}, fmt.Errorf("unable to read file: %w", err) // Wrap error for context
	}
	return Buffer{Data: data, Count: int64(len(data))}, nil
}
