package reddit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialValidation_EmptyID(t *testing.T) {
	client, err := NewClient(Credentials{ID: "", Secret: "secret1", Username: "user1", Password: "pass1"})
	require.Nil(t, client)
	require.EqualError(t, err, "credentials: client ID and secret are required")
}

func TestCredentialValidation_EmptySecret(t *testing.T) {
	client, err := NewClient(Credentials{ID: "id1", Secret: "", Username: "user1", Password: "pass1"})
	require.Nil(t, client)
	require.EqualError(t, err, "credentials: client ID and secret are required")
}

func TestCredentialValidation_UsernameWithoutPassword(t *testing.T) {
	client, err := NewClient(Credentials{ID: "id1", Secret: "secret1", Username: "user1", Password: ""})
	require.Nil(t, client)
	require.EqualError(t, err, "credentials: both username and password must be provided together")
}

func TestCredentialValidation_PasswordWithoutUsername(t *testing.T) {
	client, err := NewClient(Credentials{ID: "id1", Secret: "secret1", Username: "", Password: "pass1"})
	require.Nil(t, client)
	require.EqualError(t, err, "credentials: both username and password must be provided together")
}
