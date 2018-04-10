package influxdb

import (
	"bytes"
	"io"
	"time"
)

// PointEncoder can encode the point using the given Protocol and will return an io.Reader
// to read the encoded points.
type PointEncoder interface {
	Encode(p Protocol) (io.Reader, error)
}

// Tag is a key/value pair of strings that is indexed when inserted into a measurement.
type Tag struct {
	Key   string
	Value string
}

// Tags is a list of Tag structs. For optimal efficiency, this should be inserted
// into InfluxDB in a sorted order and should only contain unique values.
type Tags []Tag

func (a Tags) Less(i, j int) bool { return a[i].Key < a[j].Key }
func (a Tags) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Tags) Len() int           { return len(a) }

func (a Tags) String() string {
	var buf bytes.Buffer
	for i, t := range a {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(t.Key)
		buf.WriteString("=")
		buf.WriteString(t.Value)
	}
	return buf.String()
}

// Fields is a mapping of keys to field values. The values must be a float64,
// int64, uint64, string, or bool. For uint64 to work, your server must be compiled
// with uint64 support enabled.
type Fields map[string]interface{}

// Point represents a point to be written.
type Point struct {
	Name   string
	Tags   Tags
	Fields Fields
	Time   time.Time
}

// Encode will encode the point into a PointEncoder.
func (pt *Point) Encode(p Protocol) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	if err := p.Encode(buf, pt); err != nil {
		return nil, err
	}
	return buf, nil
}

// Points is a slice of points that will all be written together.
type Points []*Point

func (a Points) Encode(p Protocol) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	for _, pt := range a {
		if err := p.Encode(buf, pt); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

// PointBuffer holds a buffer of encoded points that can then be written to
// a new location.
type PointBuffer struct {
	bytes.Buffer
	p Protocol
}

// NewPointBuffer constructs a new PointBuffer.
func NewPointBuffer(p Protocol) *PointBuffer {
	return &PointBuffer{p: p}
}

// WritePoint will encode the point to the PointBuffer.
func (pb *PointBuffer) WritePoint(pt *Point) error {
	return pb.p.Encode(&pb.Buffer, pt)
}

// Encode will return a byte reader that contains the data in this PointBuffer.
func (pb *PointBuffer) Encode(p Protocol) (io.Reader, error) {
	// If the protocols are different, return an error.
	if pb.p != p {
		return nil, ErrMismatchedProtocol
	}

	// Return the buffer directly using a new bytes.Reader.
	return bytes.NewReader(pb.Buffer.Bytes()), nil
}
