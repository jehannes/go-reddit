package reddit

import (
	"encoding/json"
	"testing"
)

func TestPreviewUnmarshalling(t *testing.T) {
	// Sample JSON data with preview structure from Reddit API
	jsonData := `{
		"id": "test123",
		"title": "Test Post with Preview",
		"preview": {
			"images": [
				{
					"source": {
						"url": "https://preview.redd.it/example.jpg?width=1164&format=pjpg&auto=webp&s=abc123",
						"width": 1164,
						"height": 888
					},
					"resolutions": [
						{
							"url": "https://preview.redd.it/example.jpg?width=108&format=pjpg&auto=webp&s=def456",
							"width": 108,
							"height": 82
						},
						{
							"url": "https://preview.redd.it/example.jpg?width=216&format=pjpg&auto=webp&s=ghi789",
							"width": 216,
							"height": 165
						}
					],
					"variants": {},
					"id": "img123"
				}
			],
			"enabled": true
		}
	}`

	var post Post
	err := json.Unmarshal([]byte(jsonData), &post)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Test basic fields
	if post.ID != "test123" {
		t.Errorf("Expected ID to be 'test123', got '%s'", post.ID)
	}

	if post.Title != "Test Post with Preview" {
		t.Errorf("Expected title to be 'Test Post with Preview', got '%s'", post.Title)
	}

	// Test preview field exists
	if post.Preview == nil {
		t.Fatal("Expected preview to be present, got nil")
	}

	// Test preview enabled
	if !post.Preview.Enabled {
		t.Error("Expected preview to be enabled")
	}

	// Test preview images
	if len(post.Preview.Images) != 1 {
		t.Fatalf("Expected 1 preview image, got %d", len(post.Preview.Images))
	}

	image := post.Preview.Images[0]

	// Test source image
	expectedURL := "https://preview.redd.it/example.jpg?width=1164&format=pjpg&auto=webp&s=abc123"
	if image.Source.URL != expectedURL {
		t.Errorf("Expected source URL to be '%s', got '%s'", expectedURL, image.Source.URL)
	}

	if image.Source.Width != 1164 {
		t.Errorf("Expected source width to be 1164, got %d", image.Source.Width)
	}

	if image.Source.Height != 888 {
		t.Errorf("Expected source height to be 888, got %d", image.Source.Height)
	}

	// Test image ID
	if image.ID != "img123" {
		t.Errorf("Expected image ID to be 'img123', got '%s'", image.ID)
	}

	// Test resolutions
	if len(image.Resolutions) != 2 {
		t.Fatalf("Expected 2 resolutions, got %d", len(image.Resolutions))
	}

	// Test first resolution
	res1 := image.Resolutions[0]
	expectedRes1URL := "https://preview.redd.it/example.jpg?width=108&format=pjpg&auto=webp&s=def456"
	if res1.URL != expectedRes1URL {
		t.Errorf("Expected first resolution URL to be '%s', got '%s'", expectedRes1URL, res1.URL)
	}

	if res1.Width != 108 {
		t.Errorf("Expected first resolution width to be 108, got %d", res1.Width)
	}

	if res1.Height != 82 {
		t.Errorf("Expected first resolution height to be 82, got %d", res1.Height)
	}

	// Test second resolution
	res2 := image.Resolutions[1]
	expectedRes2URL := "https://preview.redd.it/example.jpg?width=216&format=pjpg&auto=webp&s=ghi789"
	if res2.URL != expectedRes2URL {
		t.Errorf("Expected second resolution URL to be '%s', got '%s'", expectedRes2URL, res2.URL)
	}

	if res2.Width != 216 {
		t.Errorf("Expected second resolution width to be 216, got %d", res2.Width)
	}

	if res2.Height != 165 {
		t.Errorf("Expected second resolution height to be 165, got %d", res2.Height)
	}
}
