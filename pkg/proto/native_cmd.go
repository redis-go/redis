package proto

// a command read from client connection
type nativeCmd struct {
	argv [][]byte
}

func (c *nativeCmd) Args() [][]byte {
	return c.argv
}
