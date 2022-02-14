package main

import (
	"fmt"
	"github.com/mdelillo/go-utils/ansi"
	"github.com/mdelillo/go-utils/net"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const printInterval = time.Second / 10

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url> [<url> ...]\n\n", os.Args[0])
		fmt.Printf("Example: %s http://ipv4.download.thinkbroadband.com/5MB.zip http://ipv4.download.thinkbroadband.com/20MB.zip http://ipv4.download.thinkbroadband.com/50MB.zip\n", os.Args[0])
		os.Exit(1)
	}

	if err := downloadFiles(os.Args[1:]); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func downloadFiles(urls []string) error {
	downloader := net.FileDownloader{}

	var fileDownloads []net.FileDownload
	for _, url := range urls {
		filePath := filepath.Base(url)
		fileDownloads = append(fileDownloads, net.FileDownload{
			URL:      url,
			FilePath: filePath,
		})
	}

	progresses := &downloadsProgresses{
		fileDownloads: fileDownloads,
		writer: &ansi.InPlaceWriter{
			Writer:    os.Stdout,
			LineCount: len(fileDownloads),
		},
	}

	var previousPrint time.Time
	err := downloader.DownloadFilesWithProgressUpdates(fileDownloads, func(downloadProgress net.DownloadProgress) {
		progresses.update(downloadProgress)

		now := time.Now()
		if now.Sub(previousPrint) < printInterval {
			return
		}

		progresses.print()

		previousPrint = now
	})
	if err != nil {
		return err
	}

	progresses.print()

	return nil
}

type downloadsProgresses struct {
	fileDownloads []net.FileDownload
	writer        io.Writer
	progresses    map[net.FileDownload]net.DownloadProgress
}

func (d *downloadsProgresses) update(downloadProgress net.DownloadProgress) {
	if d.progresses == nil {
		d.progresses = map[net.FileDownload]net.DownloadProgress{}
	}

	d.progresses[downloadProgress.FileDownload] = downloadProgress
}

func (d *downloadsProgresses) print() {
	var output string

	for _, fileDownload := range d.fileDownloads {
		progress := d.progresses[fileDownload]

		if progress.TotalBytes > 0 && progress.DownloadedBytes == progress.TotalBytes {
			var precision time.Duration
			if progress.DownloadTime < time.Second {
				precision = time.Millisecond
			} else {
				precision = 10 * time.Millisecond
			}

			output += fmt.Sprintf("Downloaded %s (%s in %s)\n",
				fileDownload.FilePath,
				formatFileSize(float64(progress.TotalBytes)),
				progress.DownloadTime.Round(precision),
			)

			continue
		}

		if progress.DownloadedBytes == 0 {
			output += fmt.Sprintf("%s (%s)\n", filepath.Base(fileDownload.FilePath), formatFileSize(float64(progress.TotalBytes)))
			continue
		}

		var percentComplete int
		if progress.TotalBytes != 0 {
			percentComplete = int(progress.DownloadedBytes * 100 / progress.TotalBytes)
		}

		progressBar := fmt.Sprintf("[%s%s]",
			strings.Repeat("#", percentComplete/2),
			strings.Repeat(" ", 50-percentComplete/2),
		)

		remainingDownloadTime := (time.Duration(float64(progress.TotalBytes-progress.DownloadedBytes)/progress.AverageBytesPerMicrosecond) + 999*time.Microsecond) * time.Microsecond

		output += fmt.Sprintf("Downloading %s: %s  %s/%s (%s remaining)\n",
			filepath.Base(fileDownload.FilePath),
			progressBar,
			formatFileSize(float64(progress.DownloadedBytes)),
			formatFileSize(float64(progress.TotalBytes)),
			remainingDownloadTime.Truncate(time.Second),
		)
	}

	_, _ = fmt.Fprint(d.writer, output)
}

func formatFileSize(bytes float64) string {
	oneKB := math.Pow(2, 10)
	oneMB := math.Pow(2, 20)
	oneGB := math.Pow(2, 30)
	oneTB := math.Pow(2, 40)

	var output string
	if bytes < oneKB {
		output = fmt.Sprintf("%.0fB", bytes)
	} else if bytes < oneMB {
		output = fmt.Sprintf("%.2fKB", bytes/oneKB)
	} else if bytes < oneGB {
		output = fmt.Sprintf("%.2fMB", bytes/oneMB)
	} else if bytes < oneTB {
		output = fmt.Sprintf("%.2fGB", bytes/oneGB)
	} else {
		output = fmt.Sprintf("%.2fTB", bytes/oneTB)
	}

	return strings.Replace(output, ".00", "", 1)
}
