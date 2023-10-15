package h3

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

// Frame types
const (
	FRAME_DATA                = 0x00
	FRAME_HEADERS             = 0x01
	FRAME_CANCEL_PUSH         = 0x03
	FRAME_SETTINGS            = 0x04
	FRAME_PUSH_PROMISE        = 0x05
	FRAME_GOAWAY              = 0x07
	FRAME_MAX_PUSH_ID         = 0x0D
	FRAME_WEBTRANSPORT_STREAM = 0x41
)

// HTTP/3 frame
type Frame struct {
	Type      uint64
	SessionID uint64
	Length    uint64
	Data      []byte
}

func (f *Frame) Read(r io.Reader) error {
	qr := quicvarint.NewReader(r)
	t, err := quicvarint.Read(qr)
	if err != nil {
		return err
	}
	l, err := quicvarint.Read(qr)
	if err != nil {
		return err
	}

	f.Type = t

	// For most (but not all) frame types, l is the data length
	switch t {
	case FRAME_WEBTRANSPORT_STREAM:
		f.Length = 0
		f.SessionID = l
		f.Data = []byte{}
		return nil
	default:
		f.Length = l
		f.Data = make([]byte, l)
		_, err := r.Read(f.Data)
		return err
	}
}

func (f *Frame) Write(w io.Writer) (int, error) {
	buf := &bytes.Buffer{}

	buf.Write(quicvarint.Append(nil, f.Type))
	if f.Type == FRAME_WEBTRANSPORT_STREAM {
		buf.Write(quicvarint.Append(nil, f.SessionID))
	} else {
		buf.Write(quicvarint.Append(nil, f.Length))
	}
	buf.Write(f.Data)

	return w.Write(buf.Bytes())
}
