package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const cloudinaryFolder = "roamify"

var (
	cloudinaryServiceInstance *CloudinaryService
	cloudinaryOnce            sync.Once
	cloudinaryInitError       error
)

type CloudinaryService struct {
	client *cloudinary.Cloudinary
	folder string
}

type CloudinaryUploadResult struct {
	SecureURL string
	PublicID  string
}

func GetCloudinaryService() (*CloudinaryService, error) {
	cloudinaryOnce.Do(func() {
		cloudinaryServiceInstance, cloudinaryInitError = newCloudinaryService()
	})
	return cloudinaryServiceInstance, cloudinaryInitError
}

func newCloudinaryService() (*CloudinaryService, error) {
	cloudinaryURL := strings.TrimSpace(os.Getenv("CLOUDINARY_URL"))
	if cloudinaryURL == "" {
		cloudName := strings.TrimSpace(os.Getenv("CLOUDINARY_CLOUD_NAME"))
		apiKey := strings.TrimSpace(os.Getenv("CLOUDINARY_API_KEY"))
		apiSecret := strings.TrimSpace(os.Getenv("CLOUDINARY_API_SECRET"))
		if cloudName == "" || apiKey == "" || apiSecret == "" {
			return nil, errors.New("CLOUDINARY_URL or CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET must be set")
		}
		cloudinaryURL = fmt.Sprintf("cloudinary://%s:%s@%s", apiKey, apiSecret, cloudName)
	}

	client, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("initialize cloudinary client: %w", err)
	}

	return &CloudinaryService{client: client, folder: cloudinaryFolder}, nil
}

func (s *CloudinaryService) Upload(ctx context.Context, file io.Reader, filename string) (*CloudinaryUploadResult, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("cloudinary client is not configured")
	}
	if file == nil {
		return nil, errors.New("file reader is required")
	}

	params := uploader.UploadParams{
		Folder:       s.folder,
		ResourceType: "image",
	}

	result, err := s.client.Upload.Upload(ctx, file, params)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &CloudinaryUploadResult{
		SecureURL: result.SecureURL,
		PublicID:  result.PublicID,
	}, nil
}
