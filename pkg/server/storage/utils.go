package storage

type fakeReadSeeker struct{}

func (f *fakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *fakeReadSeeker) Read(p []byte) (int, error) {
	return 0, nil
}
