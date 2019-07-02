package proto

type ProtocolError struct {
	msg string
}

func (p *ProtocolError) Error() string {
	return p.msg
}
