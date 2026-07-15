package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"hotel_lobby/internal/models"
	"hotel_lobby/internal/providers/cloudinary"
	"hotel_lobby/internal/repositories"

	"github.com/google/uuid"
)

// ImageService handles image uploads and deletion via Cloudinary.
type ImageService struct {
	cloudinary *cloudinary.Client
	repo       repositories.RoomImageRepository
}

func NewImageService(cld *cloudinary.Client, repo repositories.RoomImageRepository) *ImageService {
	return &ImageService{cloudinary: cld, repo: repo}
}

// Upload accepts a multipart.File and stores the image.
// isPrimary and sortOrder are optional (zero values mean "not set").
func (s *ImageService) Upload(ctx context.Context, roomID uuid.UUID, file multipart.File, isPrimary bool, sortOrder int) (*models.RoomImage, error) {
	uploadResp, err := s.cloudinary.UploadImage(ctx, file, "rooms")
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	img := &models.RoomImage{
		ID:        uuid.New(),
		RoomID:    roomID,
		URL:       uploadResp.SecureURL,
		IsPrimary: isPrimary,
		SortOrder: sortOrder,
	}

	if err := s.repo.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("failed to save room image: %w", err)
	}

	return img, nil
}

func (s *ImageService) Delete(ctx context.Context, imageID uuid.UUID) error {
	return s.repo.Delete(ctx, imageID)
}

func (s *ImageService) Reorder(ctx context.Context, roomID uuid.UUID, imageIDs []uuid.UUID) error {
	return s.repo.Reorder(ctx, roomID, imageIDs)
}