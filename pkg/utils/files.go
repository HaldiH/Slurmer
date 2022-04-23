package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func CopyDirectory(srcDir, dstDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}
		if err := CopyFile(path.Join(srcDir, f.Name()), path.Join(dstDir, f.Name())); err != nil {
			return err
		}
	}
	return nil
}
