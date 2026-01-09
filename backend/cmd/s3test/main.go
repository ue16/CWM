package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rhnvrm/simples3"
)

func main() {
	fmt.Println("=== S3 Pre-signed URL Test ===")
	fmt.Println()

	// Initialize S3 client
	s3Client := simples3.New(
		"us-east-1",
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	)
	s3Client.SetEndpoint("http://localhost:9000")

	bucketName := "test-bucket"
	objectName := "test-file.txt"
	file, _ := os.Open("Dockerfile")
	error_count := 0

	// Test 1: Upload a file directly
	fmt.Println("Test 1: Uploading file to S3 directly...")
	uploadResp, err := s3Client.FilePut(simples3.UploadInput{
		Bucket:      bucketName,
		ObjectKey:   objectName,
		ContentType: "text/plain",
		Body:        file,
	})
	if err != nil {
		log.Printf("Upload failed: %v\n", err)
		error_count += 1
	} else {
		fmt.Printf("✓ File uploaded successfully\n")
		fmt.Printf("  Response: %+v\n", uploadResp)
	}
	fmt.Println()

	// Test 2: Use a pre-signed URL for GET
	fmt.Println("Test 2: Generating pre-signed URL for GET...")
	getPresignedURL := s3Client.GeneratePresignedURL(simples3.PresignedInput{
		Bucket:        bucketName,
		ObjectKey:     objectName,
		Method:        "GET",
		ExpirySeconds: int(time.Hour.Seconds()),
	})
	fmt.Printf("✓ GET Pre-signed URL generated\n")
	fmt.Printf("  URL: %s\n", getPresignedURL)
	fmt.Println()

	if err := testPresignedURLDownload(getPresignedURL); err != nil {
		log.Printf("Download test failed: %v\n", err)
		error_count += 1
	} else {
		fmt.Printf("✓ File downloaded successfully using pre-signed URL\n")
	}
	fmt.Println()

	// Test 3: Use a pre-signed URL for PUT
	fmt.Println("Test 3: Generating pre-signed URL for PUT...")
	putPresignedURL := s3Client.GeneratePresignedURL(simples3.PresignedInput{
		Bucket:        bucketName,
		ObjectKey:     "uploaded-via-presigned.txt",
		Method:        "PUT",
		ExpirySeconds: int(time.Hour.Seconds()),
	})
	fmt.Printf("✓ PUT Pre-signed URL generated\n")
	fmt.Printf("  URL: %s\n", putPresignedURL)
	fmt.Println()

	if err := testPresignedURLUpload(putPresignedURL, "Test content via presigned PUT"); err != nil {
		log.Printf("Upload via presigned URL failed: %v\n", err)
		error_count += 1
	} else {
		fmt.Printf("✓ File uploaded successfully using pre-signed PUT URL\n")
	}
	fmt.Println()

	// Test 4: Delete file using pre-signed URL with HTTP DELETE
	fmt.Println("Test 4: Generating pre-signed URL for DELETE...")
	deletePresignedURL := s3Client.GeneratePresignedURL(simples3.PresignedInput{
		Bucket:        bucketName,
		ObjectKey:     objectName,
		Method:        "DELETE",
		ExpirySeconds: int(time.Hour.Seconds()),
	})
	fmt.Printf("✓ DELETE Pre-signed URL generated\n")
	fmt.Printf("  URL: %s\n", deletePresignedURL)
	fmt.Println()

	if err := testPresignedURLDelete(deletePresignedURL); err != nil {
		log.Printf("Delete test failed: %v\n", err)
		error_count += 1
	} else {
		fmt.Printf("✓ File deleted successfully using pre-signed URL\n")
	}

	// Test 5: Get file info with a HEAD request
	details, err := s3Client.FileDetails(simples3.DetailsInput{
		Bucket:    "test-bucket",
		ObjectKey: "uploaded-via-presigned.txt",
	})
	if err != nil {
		log.Printf("File details test failed: %v\n", err)
		error_count += 1
	} else {
		fmt.Printf("✓ File details retrieved successfully using pre-signed URL\n")
		fmt.Printf("%v", details)
	}

	fmt.Println()
	if error_count == 0 {
		fmt.Println("=== All tests completed successfully ===")
		fmt.Println("Pre-signed URLs work correctly!")
	} else {
		fmt.Println("=== Some tests failed ===")
		fmt.Println("Pre-signed URLs may not work correctly!")
		fmt.Println("Read the output above and S3 server logs to learn more.")
	}
}

func testPresignedURLDownload(url string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("  Downloaded content of length %d bytes\n", len(body))
	return nil
}

func testPresignedURLUpload(url string, content string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("PUT request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func testPresignedURLDelete(url string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
