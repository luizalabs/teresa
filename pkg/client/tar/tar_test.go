package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func extractTar(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	names := []string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, hdr.Name)
	}
	sort.Strings(names)

	return names, nil
}

func tree(dir string) ([]string, error) {
	var names []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		basePath := fmt.Sprintf("%s%c", filepath.Clean(dir), filepath.Separator)
		name := strings.Replace(path, basePath, "", 1)
		names = append(names, name)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

func TestCreateTempSuccess(t *testing.T) {
	tmp, err := CreateTemp("testdata/create", "test", []string{})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp)

	names, err := extractTar(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Errorf("want 2; got %d", len(names))
	}

	if names[0] != "dir/file2.txt" {
		t.Errorf("want dir/file2.txt; got %s", names[0])
	}

	if names[1] != "file1.txt" {
		t.Errorf("want file1.txt; got %s", names[1])
	}
}

func TestCreateTempIgnorePatternsSuccess(t *testing.T) {
	var testCases = []struct {
		patterns      []string
		expectedFiles []string
	}{
		{[]string{"dir/"}, []string{"file1.txt"}},
		{[]string{"dir"}, []string{"file1.txt"}},
		{[]string{"dir/file2.txt"}, []string{"file1.txt"}},
		{[]string{"file2.txt"}, []string{"file1.txt"}},
		{[]string{"file1.txt"}, []string{"dir/file2.txt"}},
		{[]string{"file1.txt/"}, []string{"dir/file2.txt", "file1.txt"}},
		{[]string{"*.txt"}, []string{}},
		{[]string{"*.sqlite"}, []string{"dir/file2.txt", "file1.txt"}},
	}

	for _, tc := range testCases {
		tmp, err := CreateTemp("testdata/create", "test", tc.patterns)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmp)

		names, err := extractTar(tmp)
		if err != nil {
			t.Fatal(err)
		}

		if len(names) != len(tc.expectedFiles) {
			t.Errorf("expected %d; got %d", len(tc.expectedFiles), len(names))
		}

		for i := range names {
			if names[i] != tc.expectedFiles[i] {
				t.Errorf("expected %s, got %s", tc.expectedFiles[i], names[i])
			}
		}
	}
}

func TestCreateTempFail(t *testing.T) {
	if _, err := CreateTemp("no such dir", "test", []string{}); err == nil {
		t.Error("want error; got nil")
	}
}

func TestExtractToTempOK(t *testing.T) {
	tmp, err := ExtractToTemp("testdata/test.tgz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	names, err := tree(tmp)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)

	if len(names) != 2 {
		t.Errorf("want 2; got %d", len(names))
	}

	if names[0] != "dir/file2.txt" {
		t.Errorf("want dir/file2.txt; got %s", names[0])
	}

	if names[1] != "file1.txt" {
		t.Errorf("want file1.txt; got %s", names[1])
	}
}

func TestExtractToTempFail(t *testing.T) {
	if _, err := ExtractToTemp("no such file"); err == nil {
		t.Error("want error; got nil")
	}
}
