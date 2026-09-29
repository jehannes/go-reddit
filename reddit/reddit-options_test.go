package reddit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithHTTPClient(t *testing.T) {
	_, err := NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithHTTPClient(nil))
	require.EqualError(t, err, "*http.Client: cannot be nil")

	_, err = NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithHTTPClient(&http.Client{}))
	require.NoError(t, err)
}

// TestWithHTTPClient_DoesNotMutateCaller verifies that NewClient and NewReadonlyClient
// wrap a copy of the caller's *http.Client rather than the caller's instance. Mutating
// the caller's client stacked transports when one *http.Client was shared between two
// reddit clients: the second client's requests passed through the first client's oauth2
// transport, which overwrote the Authorization header with the first client's token
// (including on the second client's own token request).
func TestWithHTTPClient_DoesNotMutateCaller(t *testing.T) {
	var mu sync.Mutex
	var apiAuth []string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		id, _, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "missing basic auth", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"token-%s","token_type":"bearer","expires_in":3600}`, id)
	})
	mux.HandleFunc("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		apiAuth = append(apiAuth, r.Header.Get("Authorization"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	originalTransport := &http.Transport{}
	shared := &http.Client{Transport: originalTransport, Timeout: 7 * time.Second}

	newShared := func(id string) *Client {
		c, err := NewClient(
			Credentials{ID: id, Secret: "secret"},
			WithHTTPClient(shared),
			WithBaseURL(server.URL),
			WithTokenURL(server.URL+"/api/v1/access_token"),
		)
		require.NoError(t, err)
		return c
	}
	c1 := newShared("id1")
	c2 := newShared("id2")

	readonly, err := NewReadonlyClient(WithHTTPClient(shared))
	require.NoError(t, err)

	// Each client authenticates as itself despite sharing the caller's *http.Client.
	for _, c := range []*Client{c1, c2} {
		req, err := c.NewRequest(http.MethodGet, "api/v1/me", nil)
		require.NoError(t, err)
		_, err = c.Do(ctx, req, nil)
		require.NoError(t, err)
	}
	require.Equal(t, []string{"Bearer token-id1", "Bearer token-id2"}, apiAuth)

	// The caller's client is untouched by all three constructors.
	require.True(t, shared.Transport == originalTransport, "caller's Transport must not be replaced")
	require.Nil(t, shared.CheckRedirect, "caller's CheckRedirect must not be set")

	// Each reddit client holds its own copy, carrying the caller's settings.
	for _, c := range []*Client{c1, c2, readonly} {
		require.False(t, c.client == shared, "client must hold a copy, not the caller's instance")
		require.Equal(t, 7*time.Second, c.client.Timeout)
	}
}

func TestWithUserAgent(t *testing.T) {
	c, err := NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithUserAgent("test"))
	require.NoError(t, err)
	require.Equal(t, "test", c.UserAgent())

	c, err = NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithUserAgent(""))
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("golang:%s:v%s", libraryName, libraryVersion), c.UserAgent())
}

func TestWithBaseURL(t *testing.T) {
	c, err := NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithBaseURL(":"))
	urlErr, ok := err.(*url.Error)
	require.True(t, ok)
	require.Equal(t, "parse", urlErr.Op)

	baseURL := "http://localhost:8080"
	c, err = NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithBaseURL(baseURL))
	require.NoError(t, err)
	require.Equal(t, baseURL, c.BaseURL.String())
}

func TestWithTokenURL(t *testing.T) {
	c, err := NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithTokenURL(":"))
	urlErr, ok := err.(*url.Error)
	require.True(t, ok)
	require.Equal(t, "parse", urlErr.Op)

	tokenURL := "http://localhost:8080/api/v1/access_token"
	c, err = NewClient(Credentials{ID: "id1", Secret: "secret1"}, WithTokenURL(tokenURL))
	require.NoError(t, err)
	require.Equal(t, tokenURL, c.TokenURL.String())
}

func TestFromEnv(t *testing.T) {
	os.Setenv("GO_REDDIT_CLIENT_ID", "id1")
	defer os.Unsetenv("GO_REDDIT_CLIENT_ID")

	os.Setenv("GO_REDDIT_CLIENT_SECRET", "secret1")
	defer os.Unsetenv("GO_REDDIT_CLIENT_SECRET")

	os.Setenv("GO_REDDIT_CLIENT_USERNAME", "username1")
	defer os.Unsetenv("GO_REDDIT_CLIENT_USERNAME")

	os.Setenv("GO_REDDIT_CLIENT_PASSWORD", "password1")
	defer os.Unsetenv("GO_REDDIT_CLIENT_PASSWORD")

	c, err := NewClient(Credentials{ID: "id1", Secret: "secret1"}, FromEnv)
	require.NoError(t, err)
	require.Equal(t, "id1", c.ID)
	require.Equal(t, "secret1", c.Secret)
	require.Equal(t, "username1", c.Username)
	require.Equal(t, "password1", c.Password)
}
