package reddit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostPinned(t *testing.T) {
	for _, test := range []struct {
		name     string
		jsonData string
		expected bool
	}{
		{
			name: "pinned to profile",
			jsonData: `{
				"kind": "t3",
				"data": {
					"id": "pinnedtest",
					"title": "Pinned To Profile",
					"pinned": true,
					"stickied": false
				}
			}`,
			expected: true,
		},
		{
			name: "not pinned",
			jsonData: `{
				"kind": "t3",
				"data": {
					"id": "pinnedtest",
					"title": "Not Pinned",
					"pinned": false,
					"stickied": true
				}
			}`,
			expected: false,
		},
		{
			name: "field absent",
			jsonData: `{
				"kind": "t3",
				"data": {
					"id": "pinnedtest",
					"title": "No Pinned Field"
				}
			}`,
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var thing thing
			err := json.Unmarshal([]byte(test.jsonData), &thing)
			require.NoError(t, err)

			post, ok := thing.Data.(*Post)
			require.True(t, ok, "expected Post data")
			require.Equal(t, test.expected, post.Pinned)
		})
	}
}

// TestPostPinnedIndependentOfStickied guards against Pinned and Stickied being
// wired to the same JSON key: a subreddit sticky is not a profile pin.
func TestPostPinnedIndependentOfStickied(t *testing.T) {
	jsonData := `{
		"kind": "t3",
		"data": {
			"id": "pinnedtest",
			"pinned": true,
			"stickied": false
		}
	}`

	var thing thing
	err := json.Unmarshal([]byte(jsonData), &thing)
	require.NoError(t, err)

	post, ok := thing.Data.(*Post)
	require.True(t, ok, "expected Post data")
	require.True(t, post.Pinned)
	require.False(t, post.Stickied)
}
