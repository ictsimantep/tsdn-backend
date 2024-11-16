package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/minio/minio-go/v7"
)

// MinioService manages MinIO-related operations
type MinioService struct {
	client     *minio.Client
	alias      string
	endpoint   string
	accessKey  string
	secretKey  string
	bucketName string
}

// Modify NewMinioService to accept *minio.Client as a parameter
func NewMinioService(client *minio.Client) *MinioService {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Add protocol based on SSL setting
	if useSSL {
		endpoint = "https://" + endpoint
	} else {
		endpoint = "http://" + endpoint
	}

	return &MinioService{
		client:     client,
		alias:      os.Getenv("MINIO_ALIAS"),
		endpoint:   endpoint,
		accessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		secretKey:  os.Getenv("MINIO_SECRET_KEY"),
		bucketName: os.Getenv("MINIO_BUCKET"),
	}
}

// SetFilePublicRead sets a public-read policy on a file using `mc`
func (m *MinioService) SetFilePublicRead(filePath string) error {
	// Step 1: Configure MinIO alias
	cmd := exec.Command("mc", "alias", "set", m.alias, m.endpoint, m.accessKey, m.secretKey)
	var cmdOut, cmdErr bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set mc alias: %v, stderr: %s", err, cmdErr.String())
	}

	// Step 2: Set the public-read policy
	policyCmd := exec.Command("mc", "anonymous", "set", "public", fmt.Sprintf("%s/%s/%s", m.alias, m.bucketName, filePath))
	policyCmd.Stdout = &cmdOut
	policyCmd.Stderr = &cmdErr

	if err := policyCmd.Run(); err != nil {
		return fmt.Errorf("failed to set public-read policy: %v, stderr: %s", err, cmdErr.String())
	}

	return nil
}
