/*
NOTE: The file's contents are a fork of
https://github.com/secmask/go-redisproto/blob/37fbcca3f27bb291d6918e4508fe201d4f180c41/parser.go
*/
package proto

import (
	"bytes"
	"errors"
	"io"
)

var (
	ExpectNumber   = &ProtocolError{msg: "Expect Number"}
	ExpectNewLine  = &ProtocolError{msg: "Expect Newline"}
	ExpectTypeChar = &ProtocolError{msg: "Expect TypeChar"}

	InvalidNumArg   = errors.New("TooManyArg")
	InvalidBulkSize = errors.New("InvalidBulkSize")
	LineTooLong     = errors.New("LineTooLong")

	ReadBufferInitSize = 1 << 16
	MaxNumArg          = 20
	MaxBulkSize        = 1 << 16
	MaxTelnetLine      = 1 << 10
	spaceSlice         = []byte{' '}
	emptyBulk          = [0]byte{}
)

type Reader struct {
	rd         io.Reader
	buf        []byte
	parsePos   int
	writeIndex int
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{rd: reader, buf: make([]byte, ReadBufferInitSize)}
}

func (r *Reader) ReadCommands() <-chan *nativeCmd {
	cmds := make(chan *nativeCmd)
	go func() {
		for cmd, err := r.readCommand(); err == nil; cmd, err = r.readCommand() {
			cmds <- cmd
		}
		close(cmds)
	}()
	return cmds
}

// ensure that we have enough space for writing 'req' byte
func (r *Reader) requestSpace(req int) {
	ccap := cap(r.buf)
	if r.writeIndex+req > ccap {
		newBuff := make([]byte, max(ccap*2, ccap+req+ReadBufferInitSize))
		copy(newBuff, r.buf)
		r.buf = newBuff
	}
}
func (r *Reader) readSome(min int) error {
	r.requestSpace(min)
	nr, err := io.ReadAtLeast(r.rd, r.buf[r.writeIndex:], min)
	if err != nil {
		return err
	}
	r.writeIndex += nr
	return nil
}

// check for at least 'num' byte available in buf to use, wait if need
func (r *Reader) requireNBytes(num int) error {
	a := r.writeIndex - r.parsePos
	if a >= num {
		return nil
	}
	if err := r.readSome(num - a); err != nil {
		return err
	}
	return nil
}
func (r *Reader) readNumber() (int, error) {
	var neg bool
	err := r.requireNBytes(1)
	if err != nil {
		return 0, err
	}
	switch r.buf[r.parsePos] {
	case '-':
		neg = true
		r.parsePos++
		break
	case '+':
		neg = false
		r.parsePos++
		break
	}
	var num uint64
	startPos := r.parsePos
OUTER:
	for {
		for i := r.parsePos; i < r.writeIndex; i++ {
			c := r.buf[r.parsePos]
			if c >= '0' && c <= '9' {
				num = num*10 + uint64(c-'0')
				r.parsePos++
			} else {
				break OUTER
			}
		}
		if r.parsePos == r.writeIndex {
			if e := r.readSome(1); e != nil {
				return 0, e
			}
		}
	}
	if r.parsePos == startPos {
		return 0, ExpectNumber
	}
	if neg {
		return -int(num), nil
	} else {
		return int(num), nil
	}

}
func (r *Reader) discardNewLine() error {
	if e := r.requireNBytes(2); e != nil {
		return e
	}
	if r.buf[r.parsePos] == '\r' && r.buf[r.parsePos+1] == '\n' {
		r.parsePos += 2
		return nil
	}
	return ExpectNewLine
}

func (r *Reader) parseBinary() (*nativeCmd, error) {
	r.parsePos++
	numArg, err := r.readNumber()
	if err != nil {
		return nil, err
	}
	var e error
	if e = r.discardNewLine(); e != nil {
		return nil, e
	}
	switch {
	case numArg == -1:
		return nil, r.discardNewLine() // null array
	case numArg < -1:
		return nil, InvalidNumArg
	case numArg > MaxNumArg:
		return nil, InvalidNumArg
	}
	argv := make([][]byte, 0, numArg)
	for i := 0; i < numArg; i++ {
		if e = r.requireNBytes(1); e != nil {
			return nil, e
		}
		if r.buf[r.parsePos] != '$' {
			return nil, ExpectTypeChar
		}
		r.parsePos++
		var pLen int
		if pLen, e = r.readNumber(); e != nil {
			return nil, e
		}
		if e = r.discardNewLine(); e != nil {
			return nil, e
		}
		switch {
		case pLen == -1:
			argv = append(argv, nil) // null bulk
		case pLen == 0:
			argv = append(argv, emptyBulk[:]) // empty bulk
		case pLen > 0 && pLen <= MaxBulkSize:
			if e = r.requireNBytes(pLen); e != nil {
				return nil, e
			}
			argv = append(argv, r.buf[r.parsePos:(r.parsePos+pLen)])
			r.parsePos += pLen
		default:
			return nil, InvalidBulkSize
		}
		if e = r.discardNewLine(); e != nil {
			return nil, e
		}
	}
	return &nativeCmd{argv: argv}, nil
}

func (r *Reader) parseTelnet() (*nativeCmd, error) {
	nlPos := -1
	for {
		nlPos = bytes.IndexByte(r.buf, '\n')
		if nlPos == -1 {
			if e := r.readSome(1); e != nil {
				return nil, e
			}
		} else {
			break
		}
		if r.writeIndex > MaxTelnetLine {
			return nil, LineTooLong
		}
	}
	r.parsePos = r.writeIndex // we don't support pipeline in telnet mode
	return &nativeCmd{argv: bytes.Split(r.buf[:nlPos-1], spaceSlice)}, nil
}

func (r *Reader) reset() {
	r.writeIndex = 0
	r.parsePos = 0
	//r.buf = make([]byte, len(r.buf))
}

func (r *Reader) readCommand() (*nativeCmd, error) {
	// if the buf is empty, try to fetch some
	if r.parsePos >= r.writeIndex {
		if err := r.readSome(1); err != nil {
			return nil, err
		}
	}

	var cmd *nativeCmd
	var err error
	if r.buf[r.parsePos] == '*' {
		cmd, err = r.parseBinary()
	} else {
		cmd, err = r.parseTelnet()
	}
	if r.parsePos >= r.writeIndex {
		r.reset()
	}
	return cmd, err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
