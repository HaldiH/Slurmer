package utils

import (
	"archive/tar"
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func UntarDirectory(r io.Reader, dstDir string) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if header == nil {
			continue
		}

		target := filepath.Join(dstDir, header.Name)
		fileInfo := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, fileInfo.Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode())
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}
		}
	}
	return nil
}

func UnzipFile(r io.ReaderAt, size int64, destDir string) error {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}
	for _, zipFile := range zipReader.File {
		target := filepath.Join(destDir, zipFile.Name)
		if zipFile.FileInfo().IsDir() {
			if err = os.MkdirAll(target, zipFile.Mode()); err != nil {
				return err
			}
			continue
		}
		f, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		destFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, zipFile.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()
		if _, err = io.Copy(destFile, f); err != nil {
			return err
		}
	}
	return nil
}

// func TarFile(srcDir string, w io.Writer) error {
// 	tarWriter := tar.NewWriter(w)
// 	defer tarWriter.Close()

// 	oldwd, err := os.Getwd()
// 	if err != nil {
// 		return err
// 	}

// 	os.Chdir(srcDir)
// 	defer os.Chdir(oldwd)

// 	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if path == "." {
// 			return nil
// 		}

// 		fileInfo, err := d.Info()
// 		if err != nil {
// 			return err
// 		}

// 		header, err := tar.FileInfoHeader(fileInfo, "")
// 		if err != nil {
// 			return err
// 		}
// 		header.Name = path

// 		f, err := tarWriter.CreateHeader(header)
// 		if err != nil {
// 			return err
// 		}

// 		if d.IsDir() {
// 			return nil
// 		}

// 		src, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer src.Close()

// 		_, err = io.Copy(f, src)
// 		if err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

func ZipFile(srcDir string, w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	oldwd, err := os.Getwd()
	if err != nil {
		return err
	}

	os.Chdir(srcDir)
	defer os.Chdir(oldwd)

	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = path

		f, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(f, src)
		if err != nil {
			return err
		}

		return nil
	})
}
