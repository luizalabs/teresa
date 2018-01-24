package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"sort"
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

func TestCreateTempSuccess(t *testing.T) {
	tmp, err := CreateTemp("testdata", "test", []string{})
	if err != nil {
		t.Fatal(err)
	}

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
	tmp, err := CreateTemp("testdata", "test", []string{"dir"})
	if err != nil {
		t.Fatal(err)
	}

	names, err := extractTar(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 {
		t.Errorf("want 1; got %d", len(names))
	}

	if names[0] != "file1.txt" {
		t.Errorf("want file1.txt; got %s", names[0])
	}
}

func TestCreateTempFail(t *testing.T) {
	if _, err := CreateTemp("no such dir", "test", []string{}); err == nil {
		t.Error("want error; got nil")
	}
}
