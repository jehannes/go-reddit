package reddit

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBackwardCompat_PasswordGrant verifies that when all four credentials are
// populated, the client uses the password grant flow (existing oauthTokenSource).
// Validates: Requirements 3.1, 3.2, 3.3, 3.4
func TestBackwardCompat_PasswordGrant(t *testing.T) {
	var (
		gotGrantType string
		gotUsername  string
		gotPassword  string
		gotBasicAuth string
	)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		gotGrantType = r.FormValue("grant_type")
		gotUsername = r.FormValue("username")
		gotPassword = r.FormValue("password")
		gotBasicAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"test-token","token_type":"bearer","expires_in":3600,"scope":"*"}`)
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	}))
	defer apiServer.Close()

	client, err := NewClient(
		Credentials{ID: "myID", Secret: "mySecret", Username: "myUser", Password: "myPass"},
		WithBaseURL(apiServer.URL),
		WithTokenURL(tokenServer.URL),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Make an API request to trigger token acquisition
	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	_, _ = client.Do(ctx, req, nil)

	// Verify password grant was used
	require.Equal(t, "password", gotGrantType)
	require.Equal(t, "myUser", gotUsername)
	require.Equal(t, "myPass", gotPassword)

	// Verify HTTP Basic Auth header contains client ID and secret
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("myID:mySecret"))
	require.Equal(t, expectedAuth, gotBasicAuth)
}

// TestBackwardCompat_WithTokenURL verifies that WithTokenURL directs token
// requests to the specified URL for the password grant flow.
// Validates: Requirements 5.2, 5.3
func TestBackwardCompat_WithTokenURL(t *testing.T) {
	var tokenRequestReceived bool

	customTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequestReceived = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"custom-token","token_type":"bearer","expires_in":3600,"scope":"*"}`)
	}))
	defer customTokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	}))
	defer apiServer.Close()

	client, err := NewClient(
		Credentials{ID: "id1", Secret: "secret1", Username: "user1", Password: "pass1"},
		WithBaseURL(apiServer.URL),
		WithTokenURL(customTokenServer.URL),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Make an API request to trigger token acquisition
	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	_, _ = client.Do(ctx, req, nil)

	// Verify the token request went to our custom server
	require.True(t, tokenRequestReceived, "token request should have been sent to the custom token URL")
}

// TestBackwardCompat_WithHTTPClient verifies that WithHTTPClient's transport
// is used for token requests in the password grant flow.
// Validates: Requirements 6.2
func TestBackwardCompat_WithHTTPClient(t *testing.T) {
	var transportUsed bool

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"transport-token","token_type":"bearer","expires_in":3600,"scope":"*"}`)
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	}))
	defer apiServer.Close()

	// Create a recording transport that marks when it's used
	recordingTransport := &recordingRoundTripper{
		base: http.DefaultTransport,
		onRoundTrip: func() {
			transportUsed = true
		},
	}

	customHTTPClient := &http.Client{Transport: recordingTransport}

	client, err := NewClient(
		Credentials{ID: "id1", Secret: "secret1", Username: "user1", Password: "pass1"},
		WithHTTPClient(customHTTPClient),
		WithBaseURL(apiServer.URL),
		WithTokenURL(tokenServer.URL),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Make an API request to trigger token acquisition
	req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
	require.NoError(t, err)

	_, _ = client.Do(ctx, req, nil)

	// Verify the custom transport was used
	require.True(t, transportUsed, "custom HTTP client transport should have been used for requests")
}

// TestBackwardCompat_NewClientSignature verifies that NewClient compiles with
// existing call patterns — both with and without options.
// Validates: Requirements 3.2, 3.3
func TestBackwardCompat_NewClientSignature(t *testing.T) {
	// Verify NewClient(Credentials{...}) compiles (no options)
	c1, err := NewClient(Credentials{ID: "id1", Secret: "secret1", Username: "user1", Password: "pass1"})
	require.NoError(t, err)
	require.NotNil(t, c1)

	// Verify NewClient(Credentials{...}, opts...) compiles (with options)
	c2, err := NewClient(
		Credentials{ID: "id2", Secret: "secret2", Username: "user2", Password: "pass2"},
		WithUserAgent("test-agent"),
	)
	require.NoError(t, err)
	require.NotNil(t, c2)

	// Verify Credentials struct literal initialization is unchanged
	creds := Credentials{
		ID:       "id3",
		Secret:   "secret3",
		Username: "user3",
		Password: "pass3",
	}
	c3, err := NewClient(creds)
	require.NoError(t, err)
	require.NotNil(t, c3)

	// Verify client credentials (ID+Secret only) also compiles
	c4, err := NewClient(Credentials{ID: "id4", Secret: "secret4"})
	require.NoError(t, err)
	require.NotNil(t, c4)
}
