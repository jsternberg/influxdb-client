package influxdb_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/jsternberg/influxdb-client"
)

func TestTags_Sort(t *testing.T) {
	tags := []influxdb.Tag{
		{Key: "region", Value: "useast"},
		{Key: "host", Value: "server01"},
	}
	sort.Sort(influxdb.Tags(tags))

	if tags[0].Key != "host" {
		t.Errorf("have %q, want %q", tags[0].Key, "host")
	}
	if tags[0].Value != "server01" {
		t.Errorf("have %q, want %q", tags[0].Value, "server01")
	}
	if tags[1].Key != "region" {
		t.Errorf("have %q, want %q", tags[0].Key, "region")
	}
	if tags[1].Value != "useast" {
		t.Errorf("have %q, want %q", tags[0].Value, "useast")
	}
}

func TestTags_String(t *testing.T) {
	tags := []influxdb.Tag{
		{Key: "host", Value: "server01"},
		{Key: "region", Value: "useast"},
	}

	if got, want := influxdb.Tags(tags).String(), `host=server01,region=useast`; got != want {
		t.Errorf("unexpected string:\n\ngot=%#v\n\nwant=%#v\n", got, want)
	}
}

func TestPointBuffer(t *testing.T) {
	now := time.Now()

	buf := influxdb.NewPointBuffer(influxdb.DefaultProtocol)
	buf.WritePoint(&influxdb.Point{
		Name: "cpu",
		Tags: []influxdb.Tag{
			{Key: "host", Value: "server01"},
		},
		Fields: influxdb.Fields{"value": 2.0},
		Time:   now,
	})

	if got, want := buf.String(), fmt.Sprintf(`cpu,host=server01 value=2 %d
`, now.UnixNano()); got != want {
		t.Errorf("unexpected output:\n\ngot=%s\nwant=%s", got, want)
	}
}
