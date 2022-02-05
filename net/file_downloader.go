package net

import (
	"fmt"
	"io"
	"net/http"
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
	Client *http.Client
}

func NewFileDownloader() *FileDownloader {
	return &FileDownloader{
		Client: NewHTTPClient(WithTimeout(time.Hour)),
	}
}

func (d *FileDownloader) DownloadFiles(files []FileDownload) error {
	for _, file := range files {
		err := d.downloadFile(file)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", filepath.Base(file.FilePath), err)
		}
	}

	return nil
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
	if d.Client == nil {
		d.Client = NewHTTPClient(WithTimeout(10 * time.Minute))
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, fmt.Errorf("got non-2XX response: %s", resp.Status)
	}

	return resp.ContentLength, nil
}

func (d *FileDownloader) downloadFile(fileDownload FileDownload) error {
	if d.Client == nil {
		d.Client = NewHTTPClient(WithTimeout(10 * time.Minute))
	}

	req, err := http.NewRequest(http.MethodGet, fileDownload.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.Client.Do(req)
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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (d *FileDownloader) downloadFileWithCallback(fileDownload FileDownload, callback DownloadProgressCallback) error {
	if d.Client == nil {
		d.Client = NewHTTPClient(WithTimeout(10 * time.Minute))
	}

	req, err := http.NewRequest(http.MethodGet, fileDownload.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.Client.Do(req)
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
