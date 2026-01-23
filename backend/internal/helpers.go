package internal

import (
	"fmt"
	"strings"
)

func humanizeBytes(i int64) string {
	if i < 1024 {
		return fmt.Sprintf("%d B", i)
	}
	if i < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(i)/1024)
	}
	if i < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(i)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(i)/(1024*1024*1024))
}

func NewDBFromConnectionString(connStr string) (Store, error) {
	if strings.HasPrefix(connStr, "mongodb") {
		return NewMongoDB(connStr)
	} else if strings.HasPrefix(connStr, "postgres") {
		return NewPostgresDB(connStr)
	}
	return nil, fmt.Errorf("unsupported database type")
}
