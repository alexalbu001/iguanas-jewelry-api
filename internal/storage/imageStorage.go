package storage

import "context"

type ImageStorage interface {
	// Generate a URL where frontend can upload the image
	GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error)
	// Get the public URL for displaying the image
	GetImageURL(ctx context.Context, key string) (string, error)
	// Handle post-upload processing (resizing, format conversion, etc.)
	ProcessUploadComplete(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
	Store(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}
