package influxdb_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/jsternberg/influxdb-client"
)

const MaxUDPPayload = 64 * 1024

func TestUDPClient(t *testing.T) {
	// Listen on a random ephemeral port.
	saddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", saddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	addr := conn.LocalAddr()

	done := make(chan struct{})
	go func() {
		defer close(done)

		data := make([]byte, MaxUDPPayload)
		_, _, _ = conn.ReadFromUDP(data)
	}()

	// Create the UDP writer.
	client, err := influxdb.NewUDPClient(addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	now := time.Now()
	pt := influxdb.Point{
		Name:   "cpu",
		Fields: influxdb.Fields{"value": 2.0},
		Time:   now,
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for i := 0; i < 10; i++ {
		if err := client.Write(context.Background(), &pt); err != nil {
			t.Fatal(err)
		}

		select {
		case <-ticker.C:
		case <-done:
			break
		}
	}

	// Check if the packet was received.
	select {
	case <-done:
	default:
		t.Errorf("timeout while waiting for udp packet")
	}
}
