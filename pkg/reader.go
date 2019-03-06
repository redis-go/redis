package redis

import (
	"bufio"
	"io"
)

const (
	maxBytes = 4096
)

// reader reads RESP messages.
type reader struct {
	rd         *bufio.Reader
	buf        []byte
	start, end int       // buf read and write positions
	cmds       []Command // commands waiting for being read by ReadCommand()
}

func newReader(rd io.Reader) *reader {
	return &reader{
		rd:  bufio.NewReader(rd),
		buf: make([]byte, maxBytes),
	}
}
