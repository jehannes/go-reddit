package reddit

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
)

// validCredentials generates random Credentials where ID and Secret are always
// non-empty, and Username/Password are either both empty or both non-empty.
type validCredentials struct {
	Credentials
}

func (validCredentials) Generate(rand *rand.Rand, size int) reflect.Value {
	id := randomNonEmptyString(rand)
	secret := randomNonEmptyString(rand)

	var username, password string
	if rand.Intn(2) == 0 {
		// Both empty → client_credentials grant
		username = ""
		password = ""
	} else {
		// Both non-empty → password grant
		username = randomNonEmptyString(rand)
		password = randomNonEmptyString(rand)
	}

	return reflect.ValueOf(validCredentials{
		Credentials: Credentials{
			ID:       id,
			Secret:   secret,
			Username: username,
			Password: password,
		},
	})
}

func randomNonEmptyString(rand *rand.Rand) string {
	length := rand.Intn(10) + 1
	b := make([]byte, length)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}

// TestPropertyGrantTypeSelection verifies that for any valid Credentials,
// the grant type selected by NewClient is client_credentials if and only if
// Username and Password are both empty strings.
//
// **Validates: Requirements 1.1, 1.2, 3.1, 3.5**
func TestPropertyGrantTypeSelection(t *testing.T) {
	config := &quick.Config{
		Rand:     rand.New(rand.NewSource(42)),
		MaxCount: 100,
	}

	err := quick.Check(func(vc validCredentials) bool {
		// Start a test server that captures the grant_type from the token request
		var mu sync.Mutex
		var capturedGrantType string

		mux := http.NewServeMux()
		mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err == nil {
				mu.Lock()
				capturedGrantType = r.FormValue("grant_type")
				mu.Unlock()
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"test-token","token_type":"bearer","expires_in":3600}`)
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{}`)
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// Create client with the test server as token endpoint
		client, err := NewClient(vc.Credentials,
			WithTokenURL(server.URL+"/token"),
			WithBaseURL(server.URL),
		)
		if err != nil {
			// Valid credentials should never fail NewClient
			return false
		}

		// Make a request to trigger token acquisition
		req, err := client.NewRequest(http.MethodGet, "api/v1/me", nil)
		if err != nil {
			return false
		}
		// Execute the request — this triggers the OAuth transport to fetch a token
		resp, _ := client.client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		mu.Lock()
		gt := capturedGrantType
		mu.Unlock()

		// Property: grant_type is client_credentials iff Username and Password are both empty
		expectClientCredentials := vc.Username == "" && vc.Password == ""
		if expectClientCredentials {
			return gt == "client_credentials"
		}
		return gt == "password"
	}, config)

	require.NoError(t, err)
}
