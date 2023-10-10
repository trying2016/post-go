package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Reader struct {
	io.Reader
	Total    int64
	Current  int64
	callback func(f float64)
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)

	r.Current += int64(n)
	fmt.Printf("\r进度 %.2f%%", float64(r.Current*10000/r.Total)/100)
	return
}

func DownloadFile(url, filename string, progress func(f float64)) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Body.Close()
	}()
	tmpFilename := filename + ".tmp"
	f, err := os.Create(tmpFilename)
	if err != nil {
		return err
	}
	reader := &Reader{
		Reader:   r.Body,
		Total:    r.ContentLength,
		callback: progress,
	}
	n, err := io.Copy(f, reader)
	if err != nil {
		return err
	}
	if n != r.ContentLength {
		return errors.New("download failed, file length is inconsistent")
	}
	if err := os.Rename(tmpFilename, filename); err != nil {
		return err
	}
	return nil
}
