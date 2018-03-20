package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	gitignore "github.com/sabhiram/go-gitignore"
)

func addAll(tw *tar.Writer, dir string, ignorePatterns []string) error {
	ig, err := gitignore.CompileIgnoreLines(ignorePatterns...)
	if err != nil {
		return errors.Wrap(err, "compiling ignore patterns list")
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "on walk call")
		}

		basePath := filepath.ToSlash(fmt.Sprintf("%s%c", filepath.Clean(dir), filepath.Separator))
		name := filepath.ToSlash(strings.Replace(path, basePath, "", 1))

		if info.IsDir() {
			name = fmt.Sprintf("%s/", name)
		}

		if ig.MatchesPath(name) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}
		return addFile(tw, path, name, info)
	})
}

func addFile(tw *tar.Writer, path, name string, info os.FileInfo) error {
	if !info.Mode().IsRegular() {
		return nil
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return errors.Wrap(err, "failed to build tar header")
	}
	header.Name = name

	if err := tw.WriteHeader(header); err != nil {
		return errors.Wrap(err, "failed to write tar header")
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return errors.Wrap(err, "failed to copy file contents to tarball")
}

func CreateTemp(dir, prefix string, ignorePatterns []string) (string, error) {
	tmp, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}
	defer tmp.Close()

	gw := gzip.NewWriter(tmp)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addAll(tw, dir, ignorePatterns); err != nil {
		os.Remove(tmp.Name())
		return "", errors.Wrap(err, "failed to add all files")
	}

	return tmp.Name(), nil
}

func ExtractToTemp(filename string) (string, error) {
	gr, rc, err := newReadClosers(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		gr.Close()
		rc.Close()
	}()

	tmp, err := ioutil.TempDir("", "teresa")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp dir")
	}

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return tmp, nil
		}
		if err != nil {
			os.RemoveAll(tmp)
			return "", errors.Wrap(err, "tar iteration failed")
		}

		dst := filepath.Join(tmp, skipFirst(hdr.Name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(dst); err != nil {
				if err := os.MkdirAll(dst, os.FileMode(hdr.Mode)); err != nil {
					os.RemoveAll(tmp)
					return "", errors.Wrap(err, "mkdir failed")
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				os.RemoveAll(tmp)
				return "", errors.Wrap(err, "failed to create file")
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				os.RemoveAll(tmp)
				return "", errors.Wrap(err, "copy failed")
			}
		}
	}

	return tmp, nil
}

func newReadClosers(filename string) (*gzip.Reader, io.ReadCloser, error) {
	rc, err := os.Open(filename)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open failed")
	}

	gr, err := gzip.NewReader(rc)
	if err != nil {
		return nil, nil, errors.Wrap(err, "gunzip failed")
	}

	return gr, rc, nil
}

func skipFirst(path string) string {
	s := strings.Split(path, string(filepath.Separator))
	return strings.Join(s[1:], string(filepath.Separator))
}
