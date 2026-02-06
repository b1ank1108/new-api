package channel

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func TestCopyAllRequestHeaders_ExcludeKeyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	req.Header.Set("Key", "secret")
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("X-API-Key", "user-token")
	req.Header.Set("X-Goog-Api-Key", "user-token")
	req.Header.Set("X-Test", "ok")
	req.Header.Add("X-Multi", "a")
	req.Header.Add("X-Multi", "b")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Content-Length", "123")
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Host", "evil.example.com")
	c.Request = req

	target := http.Header{}
	copyAllRequestHeaders(c, &target)

	if got := target.Get("Key"); got != "" {
		t.Fatalf("expected Key header to be excluded, got %q", got)
	}
	if got := target.Get("Authorization"); got != "" {
		t.Fatalf("expected Authorization header to be excluded, got %q", got)
	}
	if got := target.Get("X-API-Key"); got != "" {
		t.Fatalf("expected X-API-Key header to be excluded, got %q", got)
	}
	if got := target.Get("X-Goog-Api-Key"); got != "" {
		t.Fatalf("expected X-Goog-Api-Key header to be excluded, got %q", got)
	}
	if got := target.Get("X-Test"); got != "ok" {
		t.Fatalf("expected X-Test header to be copied, got %q", got)
	}
	if got := target.Values("X-Multi"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("expected X-Multi header values to be copied, got %#v", got)
	}
	if got := target.Get("Connection"); got != "" {
		t.Fatalf("expected Connection header to be excluded, got %q", got)
	}
	if got := target.Get("Transfer-Encoding"); got != "" {
		t.Fatalf("expected Transfer-Encoding header to be excluded, got %q", got)
	}
	if got := target.Get("Content-Length"); got != "" {
		t.Fatalf("expected Content-Length header to be excluded, got %q", got)
	}
	if got := target.Get("Proxy-Connection"); got != "" {
		t.Fatalf("expected Proxy-Connection header to be excluded, got %q", got)
	}
	if got := target.Get("Host"); got != "" {
		t.Fatalf("expected Host header to be excluded, got %q", got)
	}
}

type passthroughHeadersTestAdaptor struct {
	targetURL string
}

func (a *passthroughHeadersTestAdaptor) Init(*relaycommon.RelayInfo) {}

func (a *passthroughHeadersTestAdaptor) GetRequestURL(*relaycommon.RelayInfo) (string, error) {
	return a.targetURL, nil
}

func (a *passthroughHeadersTestAdaptor) SetupRequestHeader(_ *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	req.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *passthroughHeadersTestAdaptor) ConvertOpenAIRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeneralOpenAIRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertRerankRequest(*gin.Context, int, dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertEmbeddingRequest(*gin.Context, *relaycommon.RelayInfo, dto.EmbeddingRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertAudioRequest(*gin.Context, *relaycommon.RelayInfo, dto.AudioRequest) (io.Reader, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertImageRequest(*gin.Context, *relaycommon.RelayInfo, dto.ImageRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertOpenAIResponsesRequest(*gin.Context, *relaycommon.RelayInfo, dto.OpenAIResponsesRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) DoRequest(*gin.Context, *relaycommon.RelayInfo, io.Reader) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) DoResponse(*gin.Context, *http.Response, *relaycommon.RelayInfo) (any, *types.NewAPIError) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) GetModelList() []string {
	return nil
}

func (a *passthroughHeadersTestAdaptor) GetChannelName() string {
	return "passthroughHeadersTestAdaptor"
}

func (a *passthroughHeadersTestAdaptor) ConvertClaudeRequest(*gin.Context, *relaycommon.RelayInfo, *dto.ClaudeRequest) (any, error) {
	return nil, nil
}

func (a *passthroughHeadersTestAdaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	return nil, nil
}

func TestDoApiRequest_PassThroughHeaders_UsesChannelKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitHttpClient()

	received := make(chan http.Header, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("X-Test", "ok")
	c.Request = req

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiKey: "channel-token",
			ChannelSetting: dto.ChannelSettings{
				PassThroughHeadersEnabled: true,
			},
		},
	}

	adaptor := &passthroughHeadersTestAdaptor{targetURL: srv.URL}
	resp, err := DoApiRequest(adaptor, c, info, bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("DoApiRequest failed: %v", err)
	}
	_ = resp.Body.Close()

	h := <-received
	if got := h.Get("X-Test"); got != "ok" {
		t.Fatalf("expected X-Test header to be passed through, got %q", got)
	}
	if got := h.Get("Authorization"); got != "Bearer channel-token" {
		t.Fatalf("expected Authorization to use channel key, got %q", got)
	}
}
