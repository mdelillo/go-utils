package net

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"time"
)

func DownloadFiles(fileDownloads []FileDownload) error {
	return NewFileDownloader().DownloadFiles(fileDownloads)
}

func DownloadFilesWithProgress(fileDownloads []FileDownload, writer io.Writer, printInterval time.Duration) error {
	downloader := NewFileDownloader()

	downloadsProgress := map[FileDownload]DownloadProgress{}

	var previousPrint time.Time

	err := downloader.DownloadFilesWithProgressUpdates(fileDownloads, func(downloadProgress DownloadProgress) {
		downloadsProgress[downloadProgress.FileDownload] = downloadProgress

		now := time.Now()

		if now.Sub(previousPrint) < printInterval {
			return
		}

		printDownloadsProgress(fileDownloads, downloadsProgress, writer)
		previousPrint = now
	})
	if err != nil {
		return err
	}

	printDownloadsProgress(fileDownloads, downloadsProgress, writer)

	return nil
}

func printDownloadsProgress(fileDownloads []FileDownload, downloadsProgress map[FileDownload]DownloadProgress, writer io.Writer) {
	_, _ = fmt.Fprint(writer, "\033[H\033[2J")
	for _, fileDownload := range fileDownloads {
		progress := downloadsProgress[fileDownload]

		if progress.TotalBytes > 0 && progress.DownloadedBytes == progress.TotalBytes {
			var precision time.Duration
			if progress.DownloadTime < time.Second {
				precision = time.Millisecond
			} else {
				precision = 10 * time.Millisecond
			}
			_, _ = fmt.Fprintf(writer, "Downloaded %s (%s in %s)\n",
				fileDownload.FilePath,
				formatFileSize(float64(progress.TotalBytes)),
				progress.DownloadTime.Round(precision),
			)
			continue
		}

		if progress.DownloadedBytes == 0 {
			_, _ = fmt.Fprintln(writer, filepath.Base(fileDownload.FilePath))
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

		_, _ = fmt.Fprintf(writer,
			"Downloading %s: %s  %s/%s (%s remaining)\n",
			filepath.Base(fileDownload.FilePath),
			progressBar,
			formatFileSize(float64(progress.DownloadedBytes)),
			formatFileSize(float64(progress.TotalBytes)),
			remainingDownloadTime.Truncate(time.Second),
		)
	}
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
