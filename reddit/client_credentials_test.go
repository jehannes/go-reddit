package reddit

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestClientCredentials_GrantTypeAndBasicAuth verifies that the client credentials flow
// sends grant_type=client_credentials with HTTP Basic Auth containing the client ID and secret.
//
// Requirements: 2.1
func TestClientCredentials_GrantTypeAndBasicAuth(t *testing.T) {
	var mu sync.Mutex
	var capturedGrantType string
	var capturedAuthHeader string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		err := r.ParseForm()
		require.NoError(t, err)

		mu.Lock()
		capturedGrantType = r.FormValue("grant_type")
		capturedAuthHeader = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"test-token-123","token_type":"bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	clientID := "my-client-id"
	clientSecret := "my-client-secret"

	client, err := NewClient(
		Credentials{ID: clientID, Secret: clientSecret},
		WithBaseURL(server.URL),
		WithTokenURL(server.URL+"/api/v1/access_token"),
	)
	require.NoError(t, err)

	// Make a request to trigger token acquisition
	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	resp, err := client.Do(ctx, req, nil)
	require.NoError(t, err)
	resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()

	// Verify grant_type=client_credentials was sent
	require.Equal(t, "client_credentials", capturedGrantType)

	// Verify HTTP Basic Auth header contains client ID and secret
	require.True(t, strings.HasPrefix(capturedAuthHeader, "Basic "))
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(capturedAuthHeader, "Basic "))
	require.NoError(t, err)
	require.Equal(t, clientID+":"+clientSecret, string(decoded))
}

// TestClientCredentials_BearerTokenOnAPIRequests verifies that after token acquisition,
// the Authorization: Bearer <token> header is attached to API requests.
//
// Requirements: 2.2, 2.3
func TestClientCredentials_BearerTokenOnAPIRequests(t *testing.T) {
	var mu sync.Mutex
	var capturedAPIAuthHeader string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"bearer-token-abc","token_type":"bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		capturedAPIAuthHeader = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(
		Credentials{ID: "id1", Secret: "secret1"},
		WithBaseURL(server.URL),
		WithTokenURL(server.URL+"/api/v1/access_token"),
	)
	require.NoError(t, err)

	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	resp, err := client.Do(ctx, req, nil)
	require.NoError(t, err)
	resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()

	// Verify the Bearer token is attached to the API request
	require.Equal(t, "Bearer bearer-token-abc", capturedAPIAuthHeader)
}

// TestClientCredentials_WithTokenURL verifies that WithTokenURL directs token requests
// to the specified URL for the client_credentials flow.
//
// Requirements: 5.1, 5.3
func TestClientCredentials_WithTokenURL(t *testing.T) {
	var tokenServerHit bool
	var mu sync.Mutex

	// Separate token server
	tokenMux := http.NewServeMux()
	tokenMux.HandleFunc("/custom/token", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		tokenServerHit = true
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"custom-token","token_type":"bearer","expires_in":3600}`)
	})
	tokenServer := httptest.NewServer(tokenMux)
	defer tokenServer.Close()

	// API server
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})
	apiServer := httptest.NewServer(apiMux)
	defer apiServer.Close()

	client, err := NewClient(
		Credentials{ID: "id1", Secret: "secret1"},
		WithBaseURL(apiServer.URL),
		WithTokenURL(tokenServer.URL+"/custom/token"),
	)
	require.NoError(t, err)

	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	resp, err := client.Do(ctx, req, nil)
	require.NoError(t, err)
	resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()

	// Verify the token request went to the custom token server
	require.True(t, tokenServerHit, "token request should have been sent to the custom token URL")
}

// TestClientCredentials_WithHTTPClient verifies that WithHTTPClient's transport is used
// for token requests in the client_credentials flow.
//
// Requirements: 6.1, 6.3
func TestClientCredentials_WithHTTPClient(t *testing.T) {
	var mu sync.Mutex
	var transportUsed bool

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"transport-token","token_type":"bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Custom transport that records when it's used
	recording := &recordingRoundTripper{
		base: http.DefaultTransport,
		onRoundTrip: func() {
			mu.Lock()
			transportUsed = true
			mu.Unlock()
		},
	}

	customHTTPClient := &http.Client{Transport: recording}

	client, err := NewClient(
		Credentials{ID: "id1", Secret: "secret1"},
		WithHTTPClient(customHTTPClient),
		WithBaseURL(server.URL),
		WithTokenURL(server.URL+"/api/v1/access_token"),
	)
	require.NoError(t, err)

	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	resp, err := client.Do(ctx, req, nil)
	require.NoError(t, err)
	resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()

	// Verify the custom transport was used
	require.True(t, transportUsed, "custom HTTP client transport should have been used for requests")
}

// recordingRoundTripper is a test helper that records when RoundTrip is called.
// Shared across test files in the reddit package.
type recordingRoundTripper struct {
	base        http.RoundTripper
	onRoundTrip func()
}

func (rt *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.onRoundTrip != nil {
		rt.onRoundTrip()
	}
	return rt.base.RoundTrip(req)
}
