package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name         string
		offset       int64
		limit        int64
		expectedFile string
	}{
		{name: "entire file", offset: 0, limit: 0, expectedFile: "out_offset0_limit0.txt"},
		{name: "limit 10", offset: 0, limit: 10, expectedFile: "out_offset0_limit10.txt"},
		{name: "limit 1000", offset: 0, limit: 1000, expectedFile: "out_offset0_limit1000.txt"},
		{name: "limit 10000", offset: 0, limit: 10000, expectedFile: "out_offset0_limit10000.txt"},
		{name: "offset 100 limit 1000", offset: 100, limit: 1000, expectedFile: "out_offset100_limit1000.txt"},
		{name: "offset 6000 limit 1000", offset: 6000, limit: 1000, expectedFile: "out_offset6000_limit1000.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toPath := filepath.Join(t.TempDir(), "out.txt")
			fromPath := filepath.Join("testdata", "input.txt")

			err := Copy(fromPath, toPath, tt.offset, tt.limit)
			require.NoError(t, err)

			actual, err := os.ReadFile(toPath)
			require.NoError(t, err)

			expected, err := os.ReadFile(filepath.Join("testdata", tt.expectedFile))
			require.NoError(t, err)

			require.Equal(t, expected, actual)
		})
	}
}

func TestCopyOffsetExceedsFileSize(t *testing.T) {
	toPath := filepath.Join(t.TempDir(), "out.txt")
	fromPath := filepath.Join("testdata", "input.txt")

	err := Copy(fromPath, toPath, 100000, 0)
	require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
}

func TestCopyUnsupportedFile(t *testing.T) {
	toPath := filepath.Join(t.TempDir(), "out.txt")

	err := Copy(t.TempDir(), toPath, 0, 0)
	require.ErrorIs(t, err, ErrUnsupportedFile)
}
