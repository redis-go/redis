package redis

import "sync/atomic"

type stats struct {
	clientCounter uint32 // number of clients since redis start
}

// each new connection creates a client which needs an id
func (s *stats) nextClientId() clientId {
	return clientId(atomic.AddUint32(&s.clientCounter, 1))
}
