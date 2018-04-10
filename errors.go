package influxdb

import "errors"

var (
	// ErrMismatchProtocol is returned when a previously encoded point is told
	// to be encoded with a different protocol.
	ErrMismatchedProtocol = errors.New("mismatched protocol")

	// ErrNoMeasurement is returned when attempting to write with no measurement name.
	ErrNoMeasurement = errors.New("no measurement name")

	// ErrNoFields is returned when attempting to write with no fields.
	ErrNoFields = errors.New("no fields")

	// ErrNoDatabase is returned when a database hasn't been specified while writing.
	ErrNoDatabase = errors.New("no database specified")

	// ErrNotInfluxDB is returned when a ping returns from a non-InfluxDB server.
	ErrNotInfluxDB = errors.New("not an influxdb server")
)
