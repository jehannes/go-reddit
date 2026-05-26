package reddit

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// invalidCredentials is a custom type that generates Credentials where:
// - ID or Secret is empty, OR
// - exactly one of Username/Password is non-empty.
// These should always be rejected by NewClient.
type invalidCredentials struct {
	Credentials
}

func (invalidCredentials) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate random non-empty strings for fields
	randStr := func() string {
		length := r.Intn(20) + 1
		b := make([]byte, length)
		for i := range b {
			b[i] = byte(r.Intn(26) + 'a')
		}
		return string(b)
	}

	var creds Credentials

	// Pick one of four invalid categories
	category := r.Intn(4)
	switch category {
	case 0:
		// Empty ID (with any other fields)
		creds.ID = ""
		creds.Secret = randStr()
		creds.Username = randStr()
		creds.Password = randStr()
	case 1:
		// Empty Secret (with any other fields)
		creds.ID = randStr()
		creds.Secret = ""
		creds.Username = randStr()
		creds.Password = randStr()
	case 2:
		// Username non-empty but Password empty
		creds.ID = randStr()
		creds.Secret = randStr()
		creds.Username = randStr()
		creds.Password = ""
	case 3:
		// Password non-empty but Username empty
		creds.ID = randStr()
		creds.Secret = randStr()
		creds.Username = ""
		creds.Password = randStr()
	}

	return reflect.ValueOf(invalidCredentials{creds})
}

// TestPropertyInvalidCredentialCombinationsRejected validates that for any Credentials where
// ID or Secret is empty, OR where exactly one of Username/Password is non-empty,
// NewClient SHALL return a nil Client and a non-nil error.
//
// **Validates: Requirements 1.3, 1.4**
func TestPropertyInvalidCredentialCombinationsRejected(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
		Rand:     rand.New(rand.NewSource(42)),
	}

	property := func(ic invalidCredentials) bool {
		client, err := NewClient(ic.Credentials)
		// Must return nil client and non-nil error
		return client == nil && err != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: invalid credential combinations must always be rejected: %v", err)
	}
}
