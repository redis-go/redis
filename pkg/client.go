package redis

type client struct {
	conn *conn
}

// newClient returns a new client.
func newClient(conn *conn) *client {
	return &client{
		conn: conn,
	}
}
