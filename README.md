# InfluxDB Client
[![GoDoc Reference](https://godoc.org/github.com/jsternberg/influxdb-client?status.svg)](https://godoc.org/github.com/jsternberg/influxdb-client) [![experimental](http://badges.github.io/stability-badges/dist/experimental.svg)](http://github.com/badges/stability-badges)

An experimental client library for InfluxDB. This client library is
designed to make working with InfluxDB easier and support the full
breadth of features within InfluxDB at the same time.

This library is unofficial. The official library is located
[here](https://github.com/influxdata/influxdb/tree/master/client/v2).

## Example

```go
package main

import (
	"context"

	"github.com/jsternberg/influxdb-client"
)

func main() {
	client := influxdb.Client{}

	ctx := influxdb.WithDB(context.Background(), "db0")

influxdb.Execute("CREATE DATABASE mydb", nil)
defer influxdb.Execute(client.Query("DROP DATABASE mydb", nil))

now := time.Now().Truncate(time.Second).Add(-99*time.Second)
pt := influxdb.Point{Name: "cpu"}
client.WriteBatch(func(w influxdb.Writer) error {
  for i := 0; i < 100; i++ {
    pt.Fields = influxdb.Value(i)
    pt.Time = now.Add(i*time.Second)
    // You are able to use the same point after every write because the
    // values are encoded immediately.
    if err := w.WritePoint(pt); err != nil {
      return err
    }
  }
})

cur, err := client.Select("SELECT mean(value) FROM cpu",
  &influxdb.QueryOptions{Database: "mydb"})
if err != nil {
	log.Fatal(err)
}
defer cur.Close()

// Read the result set from the cursor.
result, err := cur.NextSet()
if err != nil {
  log.Fatal(err)
}

for {
  series, err := cur.NextSeries()
  if err != nil {
    if err == io.EOF {
      break
    }
    log.Fatal(err)
  }
  columns := result.Columns()

  fmt.Printf("name: %v\n", series.Name())
  for {
    row, err := series.NextRow()
    if err != nil {
      if err == io.EOF {
        break
      }
      log.Fatal(err)
    }

    values := row.Values()
    for i, column := range columns {
			fmt.Printf("%v: %v\n", column, values[i])
    }
  }
}
```

## Writing Data

Writing data to a database is very simple.

```go
client := influxdb.Client{}
client.Database = "mydb"
pt := influxdb.NewPoint("cpu", influxdb.Value(2.0), time.Time{})
client.WritePoints([]influxdb.Point{pt}, nil)
```

This will write the following line over the line protocol:

```
cpu value=2.0
```

### Using custom fields

The most common field key is `value` and the
`influxdb.Value(interface{})` function exists for this common use case.
If you need to write more than one field, you can create a map of
fields very easily.

```go
client := influxdb.Client{}
client.Database = "mydb"

fields := influxdb.Fields{
	"value": 2.0,
	"total": 10.0,
}
pt := influxdb.NewPoint("cpu", fields, time.Time{})

client.WritePoints([]influxdb.Point{pt}, nil)
```

Any of the following types can be used as a field value. The real type
is in bold and any of the other supported types will be cast to the
bolded type by the client.

* float32, **float64**
* int, int32, **int64**
* **string**
* **bool**

Unsigned integers aren't supported by InfluxDB and they are not
automatically cast so that precision isn't lost.

### Writing a point with tags

A point can be written with tags very easily by passing in a list of
tags.

```go
client := influxdb.Client{}
client.Database = "mydb"

tags := []influxdb.Tag{
	{Key: "host", Value: "server01"},
  {Key: "region", Value: "useast"},
}
pt := influxdb.NewPointWithTags("cpu", tags, influxdb.Value(2.0), time.Time{})

client.Write([]influxdb.Point{pt}, nil)
```

When writing, the tags should be sorted for best write performance. The
API uses a slice instead of a map for keeping tags so the writer does
not have to create a slice and sort the tags itself everytime you write.
For best performance, try to reuse tags for multiple calls.

different types of writing
* immediate
* batched time flush