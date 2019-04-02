package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// DefaultURL is the default url for InfluxDB.
var DefaultURL = url.URL{
	Scheme: "http",
	Host:   "localhost:8086",
}

// Querier is an interface for querying for data.
type Querier interface {
	Query(ctx context.Context, query string) error
}

// Writer will write a PointEncoder to the underlying Writer.
type Writer interface {
	// Write will write the points from the encoder to the endpoint.
	Write(ctx context.Context, enc PointEncoder) error

	// Protocol returns the protocol used for this writer.
	Protocol() Protocol
}

// DefaultUserAgent contains the default user agent when none is set.
const DefaultUserAgent = "jsternberg/influxdb-client"

// Client is the HTTP client for writing data.
type Client struct {
	// Client contains the Client that will be used to make requests.
	// If this is blank, http.DefaultClient is used.
	Client *http.Client

	// URL holds the base URL for the InfluxDB HTTP service.
	// If this is blank, http://localhost:8086 is used.
	URL *url.URL

	// UserAgent sets the user agent for the client requests.
	UserAgent string

	// Protocol contains the write protocol that should be used when
	// encoding points. If this is not set, the DefaultProtocol will
	// be used.
	Protocol Protocol

	// Database is the default database that this client should use
	// for queries and writes. If this is blank, no default database
	// will be used and the database will have to be set in the context
	// using WithDB.
	Database string

	// RetentionPolicy is the default retention policy that this client
	// should use for writes. If this is blank, it will leave it up
	// to the server which retention policy to use. This option can
	// be set in the context by using WithRP. This option will do nothing
	// for queries.
	RetentionPolicy string
}

// ServerInfo contains the server information obtained from a ping.
type ServerInfo struct {
	// Version is the version returned by InfluxDB. This will either be
	// a semantic version or the string "unknown".
	Version string
}

// Ping will ping the server to check if it is working and will return an
// error if the ping was unsucessful. If the server is not an InfluxDB
// server, but responds to the /ping endpoint, this will throw an error.
func (c *Client) Ping(ctx context.Context) (ServerInfo, error) {
	req, _ := http.NewRequest("GET", c.url("/ping", nil), nil)
	req = req.WithContext(ctx)

	resp, err := c.client().Do(req)
	if err != nil {
		return ServerInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return ServerInfo{}, fmt.Errorf("could not read /ping response: %s", err)
		}
		return ServerInfo{}, errors.New(string(out))
	}

	version := resp.Header.Get("X-Influxdb-Version")
	if version == "" {
		return ServerInfo{}, ErrNotInfluxDB
	}
	return ServerInfo{
		Version: version,
	}, nil
}

// Write will write the point to the HTTP endpoint.
func (c *Client) Write(ctx context.Context, enc PointEncoder) error {
	p := c.Protocol
	if p == nil {
		p = DefaultProtocol
	}

	body, err := enc.Encode(p)
	if err != nil {
		return err
	}

	params := url.Values{}
	if db := DBFromContext(ctx); db != "" {
		params.Set("db", db)
	} else if c.Database != "" {
		params.Set("db", c.Database)
	} else {
		return ErrNoDatabase
	}
	if rp := RPFromContext(ctx); rp != "" {
		params.Set("rp", rp)
	} else if c.RetentionPolicy != "" {
		params.Set("rp", rp)
	}

	// Always set the precision to nanoseconds. It's a mess to support anything else.
	params.Set("precision", "ns")

	req, _ := http.NewRequest("POST", c.url("/write", params), body)
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", p.ContentType())

	userAgent := c.UserAgent
	if userAgent != "" {
		userAgent = DefaultUserAgent
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not read /write response: %s", err)
		}

		// Attempt to decode this as a JSON error.
		var jsonErr struct {
			Err string `json:"error"`
		}
		if err := json.Unmarshal(out, &jsonErr); err != nil {
			// An error occurred decoding the JSON so just return the direct text.
			return errors.New(string(bytes.TrimSpace(out)))
		}
		return errors.New(jsonErr.Err)
	}
	return nil
}

func (c *Client) Query(ctx context.Context, query string) error {
	return nil
}

// client returns the client to be used for requests.
func (c *Client) client() *http.Client {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	return client
}

// url returns the string url for the request. The path must begin with a slash.
func (c *Client) url(path string, params url.Values) string {
	u := DefaultURL
	if c.URL != nil {
		u = *c.URL
	}
	if strings.HasSuffix(u.Path, "/") {
		path = strings.TrimRight(u.Path, "/")
	}
	u.Path += path
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}
	return u.String()
}
