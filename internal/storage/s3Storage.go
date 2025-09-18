package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/image/draw"
)

type S3Storage struct {
	s3client  *s3.Client
	bucket    string
	region    string
	baseURL   string
	presigner *s3.PresignClient
}

func NewS3Storage(s3client *s3.Client, bucket, region, baseURL string, presigner *s3.PresignClient) *S3Storage {
	return &S3Storage{
		s3client:  s3client,
		bucket:    bucket,
		region:    region,
		baseURL:   baseURL,
		presigner: presigner,
	}
}

func (s *S3Storage) GenerateUploadURL(ctx context.Context, key, contentType string) (string, error) {
	presignedURL, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 15
	})
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}

func (s *S3Storage) GetImageURL(ctx context.Context, key string) (string, error) {
	presignedURL, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 15
	})
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}

// This delete goes through the API. It doesn't generate a presigned URL for client because its not a safe operation.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) ProcessUploadComplete(ctx context.Context, key string) error {
	data, err := s.Get(ctx, key) // Raw bytes
	if err != nil {
		return err
	}
	img, format, err := image.Decode(bytes.NewReader(data)) // Decode when needed

	resizedImg := resizeImage(img, 800, 800)

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resizedImg)
	case "webp":
		// Convert WebP to JPEG for storage
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	return s.Store(ctx, key, buf.Bytes())
}

func resizeImage(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	newWidth, newHeight := calculateFitDimensions(srcWidth, srcHeight, maxWidth, maxHeight)

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Scale with bilinear interpolation
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

func calculateFitDimensions(srcWidth, srcHeight, maxWidth, maxHeight int) (int, int) {
	if srcWidth <= maxWidth && srcHeight <= maxHeight {
		return srcWidth, srcHeight
	}

	aspectRatio := float64(srcWidth) / float64(srcHeight)

	if float64(maxWidth)/aspectRatio <= float64(maxHeight) {
		return maxWidth, int(float64(maxWidth) / aspectRatio)
	}

	return int(float64(maxHeight) * aspectRatio), maxHeight
}

func (s *S3Storage) Store(ctx context.Context, key string, data []byte) error {
	_, err := s.s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(s.getContentType(key)), // "image/jpeg" etc.
	})
	return err
}

func (s *S3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := s.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if object doesn't exist
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, fmt.Errorf("image not found: %s", key)
		}
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (s *S3Storage) getContentType(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
