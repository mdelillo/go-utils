package main

import (
	"fmt"
	"github.com/mdelillo/go-utils/net"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url> [<url> ...]\n\n", os.Args[0])
		fmt.Printf("Example: %s http://ipv4.download.thinkbroadband.com/5MB.zip http://ipv4.download.thinkbroadband.com/20MB.zip http://ipv4.download.thinkbroadband.com/50MB.zip\n", os.Args[0])
		os.Exit(1)
	}

	var fileDownloads []net.FileDownload
	for _, url := range os.Args[1:] {
		filePath := filepath.Base(url)
		fileDownloads = append(fileDownloads, net.FileDownload{
			URL:      url,
			FilePath: filePath,
		})
	}

	if err := net.DownloadFilesWithProgress(fileDownloads); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
