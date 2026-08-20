// Package stream is responsible for handling
// all the operations pertaining to raw tcp data stream
package stream

import (
	"bufio"
	"io"
	"strings"
)

type Reader struct {
	buffered *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		buffered: bufio.NewReader(r),
	}
}

func (r *Reader) ReadExact(count int) ([]byte, error) {
	buf := make([]byte, count)

	_, err := io.ReadFull(r.buffered, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (r *Reader) ReadLine() (string, error) {
	lineBytes, err := r.buffered.ReadBytes('\n')
	trimmed := strings.TrimRight(string(lineBytes), "\r\n")
	return trimmed, err
}
