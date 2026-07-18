package httpclient

import (
	"io"
)

const maxDrainBytes = 64 * 1024

// DrainAndClose discards a bounded response remainder and always closes it.
func DrainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, maxDrainBytes))
	_ = body.Close()
}
