package server

import (
	"io"
	"log"
	"strings"
)

type FilteredWriter struct {
	io.Writer
}

var msgsToFilter = []string{
	"http2Server.HandleStreams failed to receive the preface from client",
	"remote error: tls: bad certificate",
}

func (f *FilteredWriter) Write(b []byte) (int, error) {
	for _, msg := range msgsToFilter {
		if strings.Contains(string(b), msg) {
			return 0, nil
		}
	}
	return f.Writer.Write(b)
}

func NewLogger(w io.Writer) *log.Logger {
	return log.New(&FilteredWriter{Writer: w}, "gRPC: ", log.Ldate|log.Ltime)
}
