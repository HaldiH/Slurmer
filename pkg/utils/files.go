package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

func CopyFile(src, dst string, replace bool) error {
	if !replace {
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
	}
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

// CopyDirectory copies the files inside `srcDir` into `dstDir` without recursion. Skip non-regular files.
func CopyDirectory(srcDir, dstDir string, replace bool) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}
		if err := CopyFile(path.Join(srcDir, f.Name()), path.Join(dstDir, f.Name()), replace); err != nil {
			return err
		}
	}
	return nil
}
