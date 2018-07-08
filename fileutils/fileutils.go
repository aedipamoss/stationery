package fileutils

import (
	"io"
	"os"
	"path/filepath"
)

// Basename returns only the name of a file without any extension.
func Basename(file os.FileInfo) string {
	name := file.Name()
	basename := filepath.Ext(name)
	return name[0 : len(name)-len(basename)]
}

func deferClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(err)
	}
}

// CopyFiles iterates the files array and copies each one to it's destination.
func CopyFiles(files []string, src string, dest string) error {
	for _, file := range files {
		src := filepath.Join(src, file)
		dest := filepath.Join(dest, file)

		from, err := os.Open(src)
		if err != nil {
			return err
		}
		defer deferClose(from)

		to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer deferClose(to)

		_, err = io.Copy(to, from)
		if err != nil {
			return err
		}
	}

	return nil
}
