package coreapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol"
)

type fakeTTSControl struct {
	issue   func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error)
	verify  func(context.Context, string) (ttscontrol.VerifiedInstallation, error)
	reserve func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error)
}

func (f fakeTTSControl) IssueInstallationToken(ctx context.Context, ip netip.Addr, subject string) (ttscontrol.InstallationToken, error) {
	return f.issue(ctx, ip, subject)
}

func (f fakeTTSControl) VerifyInstallationToken(ctx context.Context, token string) (ttscontrol.VerifiedInstallation, error) {
	return f.verify(ctx, token)
}

func (f fakeTTSControl) ReserveChunks(ctx context.Context, installation ttscontrol.VerifiedInstallation, chunks int) (ttscontrol.DailyQuota, error) {
	return f.reserve(ctx, installation, chunks)
}

func TestTTSControlRoutesAreAbsentWhenDisabled(t *testing.T) {
	handler := NewWithTTSControl(nil, nil, nil, nil, config.Config{}, workingTTSControl())
	for _, path := range []string{"/v1/tts/installations", "/v1/tts/anonymous"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, http.StatusNotFound)
		}
	}
}

func TestCreateTTSInstallationUsesRightmostTrustedXFF(t *testing.T) {
	var gotIP netip.Addr
	api := &API{config: config.Config{TTSControl: config.TTSControlConfig{TrustedProxyHops: 1}}, ttsControl: fakeTTSControl{
		issue: func(_ context.Context, ip netip.Addr, subject string) (ttscontrol.InstallationToken, error) {
			gotIP = ip
			return ttscontrol.InstallationToken{Token: "installation-token", Installation: ttscontrol.Installation{ExpiresAt: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)}}, nil
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{}, nil
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			return ttscontrol.DailyQuota{}, nil
		},
	}}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts/installations", bytes.NewBufferString("{}"))
	req.RemoteAddr = "192.0.2.1:443"
	req.Header.Set("X-Forwarded-For", "198.51.100.17, 203.0.113.9")

	api.createTTSInstallation(recorder, req)

	if recorder.Code != http.StatusOK || gotIP.String() != "203.0.113.9" {
		t.Fatalf("status = %d, IP = %s", recorder.Code, gotIP)
	}
}

func TestClientIPDoesNotTrustLeftmostSpoofedXFF(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/tts/installations", nil)
	req.RemoteAddr = "192.0.2.1:443"
	req.Header.Set("X-Forwarded-For", "198.51.100.17, 203.0.113.9")
	ip, err := clientIP(req, 1)
	if err != nil || ip.String() != "203.0.113.9" {
		t.Fatalf("trusted IP = %s, err = %v", ip, err)
	}
	req.Header.Set("X-Forwarded-For", "not-an-ip")
	if _, err := clientIP(req, 1); err == nil {
		t.Fatal("malformed trusted XFF was accepted")
	}
}

func TestAnonymousTTSReservesRuneChunksAndStripsInstallationToken(t *testing.T) {
	var reserved int
	var upstream *http.Request
	api := &API{
		config: config.Config{TTS: config.TTSConfig{BaseURL: "https://tts.example", ServiceToken: "service-token", MaxAudioBytes: 1024}, TTSControl: config.TTSControlConfig{ChunkChars: 2, AnonymousMaxChunks: 5}},
		ttsControl: fakeTTSControl{
			issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
				return ttscontrol.InstallationToken{}, nil
			},
			verify: func(_ context.Context, token string) (ttscontrol.VerifiedInstallation, error) {
				if token != "installation-token" {
					t.Fatalf("token = %q", token)
				}
				return ttscontrol.VerifiedInstallation{InstallationID: "installation", Kind: ttscontrol.InstallationAnonymous}, nil
			},
			reserve: func(_ context.Context, _ ttscontrol.VerifiedInstallation, chunks int) (ttscontrol.DailyQuota, error) {
				reserved = chunks
				return ttscontrol.DailyQuota{}, nil
			},
		},
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			upstream = req.Clone(req.Context())
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"audio/mpeg"}}, Body: io.NopCloser(bytes.NewBufferString("audio"))}, nil
		})},
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts/anonymous?voice=anna", bytes.NewBufferString(`{"text":" абвгд "}`))
	req.Header.Set("X-TTS-Installation-Token", "installation-token")
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("X-Request-ID", "request-id")

	api.proxyAnonymousTTS(recorder, req)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "audio" || reserved != 3 {
		t.Fatalf("status = %d, audio = %q, chunks = %d", recorder.Code, recorder.Body.String(), reserved)
	}
	if upstream.Header.Get("X-TTS-Installation-Token") != "" || upstream.Header.Get("Authorization") != "" {
		t.Fatalf("client credentials forwarded: %#v", upstream.Header)
	}
	if upstream.Header.Get("X-TTS-Service-Token") != "service-token" || upstream.Header.Get("X-Request-ID") != "request-id" {
		t.Fatalf("upstream headers = %#v", upstream.Header)
	}
}

func TestAnonymousTTSRejectsInvalidTokenAndQuota(t *testing.T) {
	invalid := &API{config: config.Config{TTSControl: config.TTSControlConfig{ChunkChars: 240, AnonymousMaxChunks: 5}}, ttsControl: fakeTTSControl{
		issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
			return ttscontrol.InstallationToken{}, nil
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{}, ttscontrol.ErrTokenInvalid
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			return ttscontrol.DailyQuota{}, nil
		},
	}}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts/anonymous", bytes.NewBufferString(`{"text":"hello"}`))
	invalid.proxyAnonymousTTS(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d", recorder.Code)
	}

	quota := *invalid
	quota.ttsControl = fakeTTSControl{
		issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
			return ttscontrol.InstallationToken{}, nil
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{Kind: ttscontrol.InstallationAnonymous}, nil
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			return ttscontrol.DailyQuota{}, ttscontrol.ErrQuotaExceeded
		},
	}
	recorder = httptest.NewRecorder()
	quota.proxyAnonymousTTS(recorder, httptest.NewRequest(http.MethodPost, "/v1/tts/anonymous", bytes.NewBufferString(`{"text":"hello"}`)))
	if recorder.Code != http.StatusTooManyRequests || recorder.Header().Get("Retry-After") == "" {
		t.Fatalf("quota response = %d, retry = %q", recorder.Code, recorder.Header().Get("Retry-After"))
	}
}

func TestAnonymousTTSFailsClosedWhenControlStorageIsUnavailable(t *testing.T) {
	api := &API{config: config.Config{TTSControl: config.TTSControlConfig{ChunkChars: 240, AnonymousMaxChunks: 5}}, ttsControl: fakeTTSControl{
		issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
			return ttscontrol.InstallationToken{}, nil
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{}, errors.New("postgres unavailable")
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			return ttscontrol.DailyQuota{}, nil
		},
	}}
	recorder := httptest.NewRecorder()
	api.proxyAnonymousTTS(recorder, httptest.NewRequest(http.MethodPost, "/v1/tts/anonymous", bytes.NewBufferString(`{"text":"hello"}`)))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("storage failure status = %d", recorder.Code)
	}
}

func TestAnonymousTTSRejectsMoreThanFiveChunksBeforeReserve(t *testing.T) {
	reserved := false
	api := &API{config: config.Config{TTSControl: config.TTSControlConfig{ChunkChars: 1, AnonymousMaxChunks: 5}}, ttsControl: fakeTTSControl{
		issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
			return ttscontrol.InstallationToken{}, nil
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{Kind: ttscontrol.InstallationAnonymous}, nil
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			reserved = true
			return ttscontrol.DailyQuota{}, nil
		},
	}}
	recorder := httptest.NewRecorder()
	api.proxyAnonymousTTS(recorder, httptest.NewRequest(http.MethodPost, "/v1/tts/anonymous", bytes.NewBufferString(`{"text":"abcdef"}`)))
	if recorder.Code != http.StatusTooManyRequests || recorder.Header().Get("Retry-After") == "" || reserved {
		t.Fatalf("over-max response = %d, retry = %q, reserved = %t", recorder.Code, recorder.Header().Get("Retry-After"), reserved)
	}
}

func workingTTSControl() fakeTTSControl {
	return fakeTTSControl{
		issue: func(context.Context, netip.Addr, string) (ttscontrol.InstallationToken, error) {
			return ttscontrol.InstallationToken{}, errors.New("not called")
		},
		verify: func(context.Context, string) (ttscontrol.VerifiedInstallation, error) {
			return ttscontrol.VerifiedInstallation{}, errors.New("not called")
		},
		reserve: func(context.Context, ttscontrol.VerifiedInstallation, int) (ttscontrol.DailyQuota, error) {
			return ttscontrol.DailyQuota{}, errors.New("not called")
		},
	}
}
