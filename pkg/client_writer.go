package redis

// Writer writes a new RESP message.
type Writer interface {
	WriteNull()
}
