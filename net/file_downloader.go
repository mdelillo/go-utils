package net

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type FileDownload struct {
	URL      string
	FilePath string
}

type DownloadProgress struct {
	FileDownload               FileDownload
	TotalBytes                 int64
	DownloadedBytes            int64
	AverageBytesPerMicrosecond float64
	DownloadTime               time.Duration
}

type DownloadProgressCallback func(DownloadProgress)

type FileDownloader struct {
	Browser *Browser
}

func (d *FileDownloader) DownloadFiles(files []FileDownload) error {
	return d.DownloadFilesWithProgressUpdates(files, func(_ DownloadProgress) {})
}

func (d *FileDownloader) DownloadFilesWithProgressUpdates(fileDownloads []FileDownload, callback DownloadProgressCallback) error {
	for _, fileDownload := range fileDownloads {
		contentLength, err := d.getContentLength(fileDownload.URL)
		if err != nil {
			return fmt.Errorf("failed to get content size of %s: %w", filepath.Base(fileDownload.FilePath), err)
		}

		callback(DownloadProgress{
			FileDownload: fileDownload,
			TotalBytes:   contentLength,
		})
	}

	for _, fileDownload := range fileDownloads {
		err := d.downloadFileWithCallback(fileDownload, callback)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", filepath.Base(fileDownload.FilePath), err)
		}
	}

	return nil
}

func (d *FileDownloader) getContentLength(url string) (int64, error) {
	resp, err := d.Browser.Head(url)
	if err != nil {
		return 0, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, fmt.Errorf("got non-2XX response: %s", resp.Status)
	}

	return resp.ContentLength, nil
}

func (d *FileDownloader) downloadFileWithCallback(fileDownload FileDownload, callback DownloadProgressCallback) error {
	resp, err := d.Browser.Get(fileDownload.URL)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("got non-2XX response: %s", resp.Status)
	}

	file, err := os.OpenFile(fileDownload.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(&downloadProgressWriter{
		writer:        file,
		callback:      callback,
		fileDownload:  fileDownload,
		contentLength: resp.ContentLength,
	}, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

type downloadProgressWriter struct {
	writer            io.Writer
	callback          DownloadProgressCallback
	fileDownload      FileDownload
	contentLength     int64
	totalWrittenBytes int64
	firstWrite        time.Time
}

func (w *downloadProgressWriter) Write(bytes []byte) (int, error) {
	if w.firstWrite.IsZero() {
		w.firstWrite = time.Now()
	}

	writtenBits, err := w.writer.Write(bytes)
	if err != nil {
		return 0, err
	}

	now := time.Now()

	w.totalWrittenBytes += int64(writtenBits)

	w.callback(DownloadProgress{
		FileDownload:               w.fileDownload,
		TotalBytes:                 w.contentLength,
		DownloadedBytes:            w.totalWrittenBytes,
		AverageBytesPerMicrosecond: float64(w.totalWrittenBytes) / float64(now.Sub(w.firstWrite).Microseconds()),
		DownloadTime:               now.Sub(w.firstWrite),
	})

	return writtenBits, nil
}
