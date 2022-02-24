package h3

import (
	"bufio"
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/lucas-clemente/quic-go"
	"github.com/marten-seemann/qpack"
)

// DataStreamer lets the caller take over the stream. After a call to DataStream
// the HTTP server library will not do anything else with the connection.
//
// It becomes the caller's responsibility to manage and close the stream.
//
// After a call to DataStream, the original Request.Body must not be used.
type DataStreamer interface {
	DataStream() quic.Stream
}

type ResponseWriter struct {
	stream         quic.Stream // needed for DataStream()
	bufferedStream *bufio.Writer

	header         http.Header
	status         int // status code passed to WriteHeader
	headerWritten  bool
	dataStreamUsed bool // set when DataSteam() is called
}

func NewResponseWriter(stream quic.Stream) *ResponseWriter {
	return &ResponseWriter{
		header:         http.Header{},
		stream:         stream,
		bufferedStream: bufio.NewWriter(stream),
	}
}

func (w *ResponseWriter) Header() http.Header {
	return w.header
}

func (w *ResponseWriter) WriteHeader(status int) {
	if w.headerWritten {
		return
	}

	if status < 100 || status >= 200 {
		w.headerWritten = true
	}
	w.status = status

	var headers bytes.Buffer
	enc := qpack.NewEncoder(&headers)
	enc.WriteField(qpack.HeaderField{Name: ":status", Value: strconv.Itoa(status)})
	for k, v := range w.header {
		for index := range v {
			enc.WriteField(qpack.HeaderField{Name: strings.ToLower(k), Value: v[index]})
		}
	}

	headersFrame := Frame{Type: FRAME_HEADERS, Length: uint64(headers.Len()), Data: headers.Bytes()}
	headersFrame.Write(w.bufferedStream)
	if !w.headerWritten {
		w.Flush()
	}
}

func (w *ResponseWriter) Write(p []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(200)
	}
	if !bodyAllowedForStatus(w.status) {
		return 0, http.ErrBodyNotAllowed
	}

	dataFrame := Frame{Type: FRAME_DATA, Length: uint64(len(p)), Data: p}
	return dataFrame.Write(w.bufferedStream)
}

func (w *ResponseWriter) Flush() {
	w.bufferedStream.Flush()
}

func (w *ResponseWriter) usedDataStream() bool {
	return w.dataStreamUsed
}

func (w *ResponseWriter) DataStream() quic.Stream {
	w.dataStreamUsed = true
	w.Flush()
	return w.stream
}

// copied from http2/http2.go
// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 2616, section 4.4.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
