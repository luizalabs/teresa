package cloudprovider

type FakeK8sOperations struct {
	CloudProviderNameErr     error
	CloudProviderNameValue   string
	SetServiceAnnotationsErr error
}

func (f *FakeK8sOperations) CloudProviderName() (string, error) {
	return f.CloudProviderNameValue, f.CloudProviderNameErr
}

func (f *FakeK8sOperations) SetServiceAnnotations(namespace, service string, annotations map[string]string) error {
	return f.SetServiceAnnotationsErr
}
