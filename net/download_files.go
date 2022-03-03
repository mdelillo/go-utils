package net

import (
	"fmt"
	"github.com/mdelillo/go-utils/ansi"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DownloadFilesWithProgress(fileDownloads []FileDownload) error {
	const printInterval = time.Second / 10

	downloader := FileDownloader{}

	progresses := &downloadsProgresses{
		fileDownloads: fileDownloads,
		writer: &ansi.InPlaceWriter{
			Writer:    os.Stdout,
			LineCount: len(fileDownloads),
		},
	}

	var previousPrint time.Time
	err := downloader.DownloadFilesWithProgressUpdates(fileDownloads, func(downloadProgress DownloadProgress) {
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
	fileDownloads []FileDownload
	writer        io.Writer
	progresses    map[FileDownload]DownloadProgress
}

func (d *downloadsProgresses) update(downloadProgress DownloadProgress) {
	if d.progresses == nil {
		d.progresses = map[FileDownload]DownloadProgress{}
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
				d.formatFileSize(float64(progress.TotalBytes)),
				progress.DownloadTime.Round(precision),
			)

			continue
		}

		if progress.DownloadedBytes == 0 {
			output += fmt.Sprintf("%s (%s)\n", filepath.Base(fileDownload.FilePath), d.formatFileSize(float64(progress.TotalBytes)))
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
			d.formatFileSize(float64(progress.DownloadedBytes)),
			d.formatFileSize(float64(progress.TotalBytes)),
			remainingDownloadTime.Truncate(time.Second),
		)
	}

	_, _ = fmt.Fprint(d.writer, output)
}

func (d *downloadsProgresses) formatFileSize(bytes float64) string {
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
