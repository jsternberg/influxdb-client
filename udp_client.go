package influxdb

import (
	"context"
	"net"
)

// udpProtocol contains the protocol used for all udp connections.
var udpProtocol = LineProtocol.V1()

// UDPClient is a client for the UDP endpoint.
type UDPClient struct {
	conn net.Conn
}

// NewUDPClient will create a new UDPClient that sends points to the given endpoint.
func NewUDPClient(addr string) (*UDPClient, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}
	return &UDPClient{conn: conn}, nil
}

// Write will write the point to the UDP endpoint.
func (c *UDPClient) Write(ctx context.Context, enc PointEncoder) error {
	body, err := enc.Encode(udpProtocol)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(body.Bytes())
	return err
}

// Close closes the UDP connection.
func (c *UDPClient) Close() error {
	return c.conn.Close()
}
