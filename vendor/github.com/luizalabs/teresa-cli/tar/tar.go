package tar

import "github.com/jhoonb/archivex"

type Writer interface {
	AddFile(string, string) error
	AddAll(string) error
}

type File interface {
	Create(string) error
	Close() error
	Writer
}

type TarFile struct {
	writer *archivex.TarFile
}

func (t *TarFile) Create(name string) error {
	t.writer = new(archivex.TarFile)
	return t.writer.Create(name)
}

func (t *TarFile) AddFile(path, filename string) error {
	return t.writer.AddFileWithName(path, filename)
}

func (t *TarFile) AddAll(path string) error {
	return t.writer.AddAll(path, false)
}

func (t *TarFile) Close() error {
	return t.writer.Close()
}

func New(name string) (File, error) {
	t := new(TarFile)
	if err := t.Create(name); err != nil {
		return nil, err
	}
	return t, nil
}
