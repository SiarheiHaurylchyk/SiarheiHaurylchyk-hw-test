package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	sourceFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := fileInfo.Size()
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	bytesToCopy := fileSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	if _, err = sourceFile.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	destinationFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	progressBar := pb.Full.Start64(bytesToCopy)
	defer progressBar.Finish()

	_, err = io.CopyN(destinationFile, progressBar.NewProxyReader(sourceFile), bytesToCopy)
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}
