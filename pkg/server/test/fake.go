package test

type FakeReadSeeker struct{}

func (f *FakeReadSeeker) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *FakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
