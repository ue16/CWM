package internal

import (
	"log"
	"os"

	"github.com/rhnvrm/simples3"
)

func NewS3ClientFromEnv() *simples3.S3 {
	base := os.Getenv("AWS_BASE")
	if base == "" {
		log.Fatal("AWS_BASE env var not set")
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	if accessKey == "" {
		log.Fatal("AWS_ACCESS_KEY env var not set")
	}
	secretKey := os.Getenv("AWS_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("AWS_SECRET_KEY env var not set")
	}

	client := simples3.New("us-east-1", accessKey, secretKey)
	client.SetEndpoint(base)

	return client
}

func S3ClientWithAlternateEndpoint(orig *simples3.S3) *simples3.S3 {
	base := os.Getenv("AWS_ALTERNATE_BASE")
	if base == "" {
		log.Printf("AWS_ALTERNATE_BASE env var not set")
		return nil
	}
	client := simples3.New(orig.Region, orig.AccessKey, orig.SecretKey)
	client.SetEndpoint(base)
	return client
}
