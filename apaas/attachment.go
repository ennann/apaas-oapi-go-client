package apaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

// AttachmentService groups file and avatar operations.
type AttachmentService struct {
	client *Client
	File   *AttachmentFileService
	Avatar *AttachmentAvatarService
}

// AttachmentFileService manages file uploads, downloads and deletions.
type AttachmentFileService struct {
	client *Client
}

// AttachmentAvatarService manages avatar image upload/download.
type AttachmentAvatarService struct {
	client *Client
}

func newAttachmentService(client *Client) *AttachmentService {
	return &AttachmentService{
		client: client,
		File:   &AttachmentFileService{client: client},
		Avatar: &AttachmentAvatarService{client: client},
	}
}

// AttachmentFileUploadParams uploads a file.
type AttachmentFileUploadParams struct {
	FileName    string
	Reader      io.Reader
	ContentType string
}

// AttachmentFileDownloadParams downloads a file.
type AttachmentFileDownloadParams struct {
	FileID string
}

// AttachmentFileDeleteParams deletes a file.
type AttachmentFileDeleteParams struct {
	FileID string
}

// AttachmentAvatarUploadParams uploads an avatar image.
type AttachmentAvatarUploadParams struct {
	FileName    string
	Reader      io.Reader
	ContentType string
}

// AttachmentAvatarDownloadParams downloads an avatar image.
type AttachmentAvatarDownloadParams struct {
	ImageID string
}

// Upload uploads a file using multipart/form-data.
func (s *AttachmentFileService) Upload(ctx context.Context, params AttachmentFileUploadParams) (*APIResponse, error) {
	if params.FileName == "" {
		return nil, fmt.Errorf("file name is required")
	}
	if params.Reader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := createFormFile(writer, "file", params.FileName, params.ContentType)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, params.Reader); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalise multipart body: %w", err)
	}

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
		"Accept":       "application/json",
	}

	s.client.log(LoggerLevelInfo, "[attachment.file.upload] Uploading file")

	resp, err := s.client.doRequestRaw(ctx, http.MethodPost, "/api/attachment/v1/files", &buf, headers, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("file upload failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode file upload response: %w", err)
	}

	s.client.log(LoggerLevelDebug, "[attachment.file.upload] File uploaded: code=%s", apiResp.Code)
	return &apiResp, nil
}

// Download retrieves a file's binary content.
func (s *AttachmentFileService) Download(ctx context.Context, params AttachmentFileDownloadParams) ([]byte, error) {
	if params.FileID == "" {
		return nil, fmt.Errorf("file ID is required")
	}
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/attachment/v1/files/%s",
		url.PathEscape(params.FileID),
	)

	data, _, err := s.client.doBinary(ctx, http.MethodGet, endpoint, nil, map[string]string{
		"Accept": "application/octet-stream",
	}, true)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[attachment.file.download] File downloaded: %s", params.FileID)
	return data, nil
}

// Delete removes a file.
func (s *AttachmentFileService) Delete(ctx context.Context, params AttachmentFileDeleteParams) (*APIResponse, error) {
	if params.FileID == "" {
		return nil, fmt.Errorf("file ID is required")
	}
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v1/files/%s",
		url.PathEscape(params.FileID),
	)

	resp, err := s.client.doJSON(ctx, http.MethodDelete, endpoint, nil, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[attachment.file.delete] File deleted: %s, code=%s", params.FileID, resp.Code)
	return resp, nil
}

// Upload uploads an avatar image.
func (s *AttachmentAvatarService) Upload(ctx context.Context, params AttachmentAvatarUploadParams) (*APIResponse, error) {
	if params.FileName == "" {
		return nil, fmt.Errorf("image name is required")
	}
	if params.Reader == nil {
		return nil, fmt.Errorf("image reader is required")
	}
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := createFormFile(writer, "image", params.FileName, params.ContentType)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, params.Reader); err != nil {
		return nil, fmt.Errorf("failed to copy image content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalise multipart body: %w", err)
	}

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
		"Accept":       "application/json",
	}

	s.client.log(LoggerLevelInfo, "[attachment.avatar.upload] Uploading avatar image")

	resp, err := s.client.doRequestRaw(ctx, http.MethodPost, "/api/attachment/v1/images", &buf, headers, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("avatar upload failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode avatar upload response: %w", err)
	}

	s.client.log(LoggerLevelDebug, "[attachment.avatar.upload] Avatar image uploaded: code=%s", apiResp.Code)
	return &apiResp, nil
}

// Download retrieves an avatar image's binary content.
func (s *AttachmentAvatarService) Download(ctx context.Context, params AttachmentAvatarDownloadParams) ([]byte, error) {
	if params.ImageID == "" {
		return nil, fmt.Errorf("image ID is required")
	}
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/attachment/v1/images/%s",
		url.PathEscape(params.ImageID),
	)

	data, _, err := s.client.doBinary(ctx, http.MethodGet, endpoint, nil, map[string]string{
		"Accept": "application/octet-stream",
	}, true)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[attachment.avatar.download] Avatar image downloaded: %s", params.ImageID)
	return data, nil
}

func createFormFile(writer *multipart.Writer, fieldName, fileName, contentType string) (io.Writer, error) {
	fileName = sanitizeFileName(fileName)
	if contentType == "" {
		return writer.CreateFormFile(fieldName, fileName)
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	header.Set("Content-Type", contentType)
	return writer.CreatePart(header)
}

func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, `"`, "_")
	if len(name) == 0 {
		return "file"
	}
	return name
}
