package reddit

import (
	"encoding/json"
	"testing"
)

func TestPostMediaMetadata(t *testing.T) {
	// Sample JSON data with media_metadata
	jsonData := `{
		"kind": "t3",
		"data": {
			"id": "gallerytest",
			"title": "Test Gallery Post",
			"is_gallery": true,
			"media_metadata": {
				"abc123def456": {
					"status": "valid",
					"e": "Image",
					"m": "image/jpg",
					"p": [
						{
							"y": 108,
							"x": 108,
							"u": "https://preview.redd.it/abc123def456.jpg?width=108"
						},
						{
							"y": 216,
							"x": 216,
							"u": "https://preview.redd.it/abc123def456.jpg?width=216"
						}
					],
					"s": {
						"y": 2048,
						"x": 1536,
						"u": "https://i.redd.it/abc123def456.jpg"
					},
					"id": "abc123def456"
				},
				"xyz789ghi012": {
					"status": "valid",
					"e": "AnimatedImage",
					"m": "image/gif",
					"p": [
						{
							"y": 108,
							"x": 192,
							"u": "https://preview.redd.it/xyz789ghi012.gif?width=108"
						}
					],
					"s": {
						"y": 1080,
						"x": 1920,
						"u": "https://i.redd.it/xyz789ghi012.gif",
						"gif": "https://i.redd.it/xyz789ghi012.gif",
						"mp4": "https://v.redd.it/xyz789ghi012.mp4"
					},
					"id": "xyz789ghi012"
				}
			},
			"gallery_data": {
				"items": [
					{
						"media_id": "abc123def456",
						"id": 1,
						"caption": "First image in the gallery"
					},
					{
						"media_id": "xyz789ghi012",
						"id": 2,
						"caption": "Cool animated GIF"
					}
				]
			}
		}
	}`

	var thing thing
	err := json.Unmarshal([]byte(jsonData), &thing)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	post, ok := thing.Data.(*Post)
	if !ok {
		t.Fatal("Expected Post data")
	}

	// Test IsGallery
	if !post.IsGallery {
		t.Error("Expected post to be a gallery")
	}

	// Test MediaMetadata
	if post.MediaMetadata == nil {
		t.Fatal("Expected media metadata to be present")
	}

	if len(post.MediaMetadata) != 2 {
		t.Errorf("Expected 2 media items, got %d", len(post.MediaMetadata))
	}

	// Test direct access to MediaMetadata
	firstMedia := post.MediaMetadata["abc123def456"]
	if firstMedia == nil {
		t.Fatal("Expected to find media item abc123def456")
	}

	if firstMedia.E != "Image" {
		t.Errorf("Expected media type Image, got %s", firstMedia.E)
	}

	if firstMedia.M != "image/jpg" {
		t.Errorf("Expected MIME type image/jpg, got %s", firstMedia.M)
	}

	// Test media source URL access
	if firstMedia.S.U != "https://i.redd.it/abc123def456.jpg" {
		t.Errorf("Expected source URL https://i.redd.it/abc123def456.jpg, got %s", firstMedia.S.U)
	}

	// Test preview images
	if len(firstMedia.P) != 2 {
		t.Errorf("Expected 2 preview images, got %d", len(firstMedia.P))
	}

	if len(firstMedia.P) > 0 && firstMedia.P[0].U != "https://preview.redd.it/abc123def456.jpg?width=108" {
		t.Errorf("Expected first preview URL https://preview.redd.it/abc123def456.jpg?width=108, got %s", firstMedia.P[0].U)
	}

	// Test animated media
	animatedMedia := post.MediaMetadata["xyz789ghi012"]
	if animatedMedia == nil {
		t.Fatal("Expected to find animated media item xyz789ghi012")
	}

	if animatedMedia.E != "AnimatedImage" {
		t.Error("Expected animated media to have type AnimatedImage")
	}

	if animatedMedia.S.MP4 != "https://v.redd.it/xyz789ghi012.mp4" {
		t.Errorf("Expected MP4 URL https://v.redd.it/xyz789ghi012.mp4, got %s", animatedMedia.S.MP4)
	}

	// Test GalleryData
	if post.GalleryData == nil {
		t.Fatal("Expected gallery data to be present")
	}

	galleryItems := post.GalleryData.Items
	if len(galleryItems) != 2 {
		t.Errorf("Expected 2 gallery items, got %d", len(galleryItems))
	}

	if galleryItems[0].MediaID != "abc123def456" {
		t.Errorf("Expected first gallery item media ID abc123def456, got %s", galleryItems[0].MediaID)
	}

	if galleryItems[0].Caption != "First image in the gallery" {
		t.Errorf("Expected first gallery item caption 'First image in the gallery', got %s", galleryItems[0].Caption)
	}

	if galleryItems[1].MediaID != "xyz789ghi012" {
		t.Errorf("Expected second gallery item media ID xyz789ghi012, got %s", galleryItems[1].MediaID)
	}
}

func TestPostMediaMetadataEdgeCases(t *testing.T) {
	// Test post without media metadata
	post := &Post{}

	if post.MediaMetadata != nil {
		t.Error("Expected nil media metadata for empty post")
	}

	if post.GalleryData != nil {
		t.Error("Expected nil gallery data for empty post")
	}

	if !post.IsGallery {
		// This is expected - IsGallery should be false for non-gallery posts
	}

	// Test post with empty media metadata
	post.MediaMetadata = make(map[string]*MediaMetadataItem)

	if len(post.MediaMetadata) != 0 {
		t.Error("Expected empty media metadata map")
	}

	if post.MediaMetadata["nonexistent"] != nil {
		t.Error("Expected nil for non-existent media in empty metadata")
	}
}
