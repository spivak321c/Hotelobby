package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Client struct {
	cld *cloudinary.Cloudinary
}

func NewClient(url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("cloudinary url is required")
	}

	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &Client{cld: cld}, nil
}

type UploadResult struct {
	SecureURL string
	PublicID  string
}

// UploadImage uploads a single image from a multipart File to Cloudinary.
func (c *Client) UploadImage(ctx context.Context, file multipart.File, folder string) (*UploadResult, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &UploadResult{
		SecureURL: resp.SecureURL,
		PublicID:  resp.PublicID,
	}, nil
}

// DeleteImage removes an image from Cloudinary by its public ID.
func (c *Client) DeleteImage(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary destroy failed: %w", err)
	}

	return nil
}
