package httpclient

import (
	"bytes"
	"io"
	"testing"
)

type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func TestDrainAndClose(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), maxDrainBytes+128)
	body := &trackingBody{Reader: bytes.NewReader(payload)}

	DrainAndClose(body)

	if !body.closed {
		t.Fatal("response body was not closed")
	}
	remaining, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read remaining body: %v", err)
	}
	if len(remaining) != 128 {
		t.Fatalf("remaining bytes = %d, want 128", len(remaining))
	}
}
