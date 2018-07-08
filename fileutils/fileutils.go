package fileutils

import (
	"io"
	"os"
	"path/filepath"
)

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

func CopyFiles(files []string, src string, dest string) error {
	for _, file := range files {
		path := filepath.Join(src, file)
		src := filepath.Join(dest, file)

		from, err := os.Open(src)
		if err != nil {
			return err
		}
		defer deferClose(from)

		to, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
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
