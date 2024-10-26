package fs

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

var excludedDirs = []string{".git", "node_modules", "venv", ".upify", "dist"}
var excludedFiles = []string{"package-lock.json", "yarn.lock"}

func CopyFilesToTempDir(srcDir, destDir string) error {
	opts := copy.Options{
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Deep
		},
		Skip: func(srcinfo os.FileInfo, src string, dest string) (bool, error) {
			relPath, err := filepath.Rel(srcDir, src)
			if err != nil {
				return false, err
			}

			for _, exclude := range excludedDirs {
				if strings.HasPrefix(relPath, exclude) || relPath == exclude {
					return true, nil
				}
			}

			for _, exclude := range excludedFiles {
				if relPath == exclude || filepath.Base(relPath) == exclude {
					return true, nil
				}
			}

			return false, nil
		},
	}

	return copy.Copy(srcDir, destDir, opts)
}

func CreateZip(sourceDir string, destFile string) error {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	return os.WriteFile(destFile, buffer.Bytes(), 0o644)
}
