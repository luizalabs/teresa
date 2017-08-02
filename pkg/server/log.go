package server

import (
	"bytes"
	"io"
	"log"
)

type FilteredWriter struct {
	io.Writer
}

var msgsToFilter = [][]byte{
	[]byte("http2Server.HandleStreams failed to receive the preface from client"),
	[]byte("failed to complete security handshake from"),
}

func (f *FilteredWriter) Write(b []byte) (int, error) {
	for _, msg := range msgsToFilter {
		if bytes.Contains(b, msg) {
			return 0, nil
		}
	}
	return f.Writer.Write(b)
}

func NewLogger(w io.Writer) *log.Logger {
	return log.New(&FilteredWriter{Writer: w}, "gRPC: ", log.Ldate|log.Ltime)
}
