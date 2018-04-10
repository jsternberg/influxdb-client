package influxdb_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/jsternberg/influxdb-client"
)

func TestLineProtocol_V1(t *testing.T) {
	enc := influxdb.LineProtocol.V1()
	points := []influxdb.Point{
		{
			Name: "cpu",
			Tags: []influxdb.Tag{
				{Key: "host", Value: "server01"},
				{Key: "region", Value: "uswest"},
			},
			Fields: influxdb.Fields{"value": 2.0},
		},
		{
			Name: "cpu,m em",
			Tags: []influxdb.Tag{
				{Key: "h, =st", Value: "se,r=ve r"},
			},
			Fields: influxdb.Fields{"value": 5},
			Time:   time.Unix(1, 0),
		},
		{
			Name:   "line",
			Fields: influxdb.Fields{"region": "usw\\es\"t"},
		},
		{
			Name:   "bool",
			Fields: influxdb.Fields{"value": true},
		},
	}

	var buf bytes.Buffer
	for _, pt := range points {
		if err := enc.Encode(&buf, &pt); err != nil {
			t.Fatal(err)
		}
	}

	want := `cpu,host=server01,region=uswest value=2
cpu\,m\ em,h\,\ \=st=se\,r\=ve\ r value=5i 1000000000
line region="usw\\es\"t"
bool value=true
`
	if got := buf.String(); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func ExampleLineProtocol_V1() {
	enc := influxdb.LineProtocol.V1()
	points := []influxdb.Point{
		{
			Name: "cpu",
			Tags: []influxdb.Tag{
				{Key: "host", Value: "server01"},
				{Key: "region", Value: "uswest"},
			},
			Fields: influxdb.Fields{"value": 2.0},
		},
	}

	for _, pt := range points {
		if err := enc.Encode(os.Stdout, &pt); err != nil {
			panic(err)
		}
	}
	// Output: cpu,host=server01,region=uswest value=2
}
