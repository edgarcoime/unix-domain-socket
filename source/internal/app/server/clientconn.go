package server

type ClientConn struct{}

func NewClientConn() (*ClientConn, error) {
	clientConn := &ClientConn{}

	return clientConn, nil
}
