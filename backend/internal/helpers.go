package internal

import "fmt"

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
