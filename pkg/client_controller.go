package redis

import (
	"sync"
)

type clientController struct {
	mu sync.RWMutex
	// Active Clients. (Map used as a HashSet)
	clients map[*client]interface{}
}

// newClientController returns a new ClientController with the give options applied.
func newClientController() *clientController {
	c := &clientController{
		clients: make(map[*client]interface{}, 0),
	}
	return c
}

func (c *clientController) addClient(client *client) {
	c.mu.Lock()
	defer c.mu.Lock()
	c.clients[client] = nil
}
