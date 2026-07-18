package dialogworker

import (
	"context"
	"errors"
	"testing"
)

func TestDialogJobErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "canceled", err: context.Canceled, want: dialogJobCanceledCode},
		{name: "timeout", err: context.DeadlineExceeded, want: dialogJobTimeoutCode},
		{name: "internal", err: errors.New("private dialog text"), want: dialogJobFailedCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dialogJobErrorCode(tt.err); got != tt.want {
				t.Fatalf("dialogJobErrorCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
