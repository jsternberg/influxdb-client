package influxdb

import (
	"context"
)

const defaultBufSize = 4096

type BufferedWriter struct {
	buf PointBuffer
	wr  Writer
}

func NewBufferedWriter(w Writer) *BufferedWriter {
	return NewBufferedWriterSize(w, defaultBufSize)
}

func NewBufferedWriterSize(w Writer, size int) *BufferedWriter {
	bw := &BufferedWriter{wr: w}
	bw.buf.p = w.Protocol()
	bw.buf.Grow(size)
	return bw
}

func (bw *BufferedWriter) Write(ctx context.Context, enc PointEncoder) error {
	buf, err := enc.Encode(bw.wr.Protocol())
	if err != nil {
		return err
	}

	// See if there is enough space in the buffered writer.
	if avail := bw.buf.Cap() - bw.buf.Len(); len(buf) > avail {
		// Flush the data in the buffer before writing.
		if err := bw.wr.Write(ctx, &bw.buf); err != nil {
			return err
		}
		bw.buf.Reset()
	}

	// Write directly to the underlying writer if the buffered points is too large.
	if len(buf) >= bw.buf.Cap() {
		return bw.wr.Write(ctx, &PointBuffer{})
	}

	// Copy bytes into the buffer. We know we have enough space because of previous checks.
	copy(bw.buf[bw.n:], buf)
	bw.n += len(buf)
	return nil
}
