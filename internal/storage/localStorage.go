package storage

import (
	"context"
	"fmt"
	"sync"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
)

type LocalImageStorage struct {
	baseURL string
	images  map[string][]byte // in memory
	mu      sync.RWMutex      // maps arent thread safe
}

func NewLocalImageStorage(baseURL string) *LocalImageStorage {
	return &LocalImageStorage{
		baseURL: baseURL,
		images:  make(map[string][]byte),
	}
}

func (l *LocalImageStorage) GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	// Return URL pointing to server's upload endpoint
	return fmt.Sprintf("%s/api/v1/admin/uploads/%s", l.baseURL, key), nil
}

func (l *LocalImageStorage) GetImageURL(ctx context.Context, key string) (string, error) { // key = filepath
	// Return URL pointing to server's serve endpoint
	return fmt.Sprintf("%s/images/%s", l.baseURL, key), nil
}

func (l *LocalImageStorage) Delete(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, exists := l.images[key]; !exists {
		return fmt.Errorf("image not found: %s", key)
	}
	delete(l.images, key)
	return nil
}

func (l *LocalImageStorage) ProcessUploadComplete(ctx context.Context, key string) error {
	l.mu.RLock()
	_, exists := l.images[key]
	l.mu.RUnlock()

	if !exists {
		return &customerrors.ErrProductImageNotFound
	}

	// For local development, just confirm it exists
	// In production, this would resize/optimize the image
	return nil

}

func (l *LocalImageStorage) Store(ctx context.Context, key string, data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.images[key] = data
	return nil
}

func (l *LocalImageStorage) Get(ctx context.Context, key string) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if _, exists := l.images[key]; !exists {
		return nil, &customerrors.ErrProductImageNotFound
	}

	return l.images[key], nil
}
