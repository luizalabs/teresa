package cloudprovider

type FakeK8sOperations struct {
	CloudProviderNameErr     error
	CloudProviderNameValue   string
	SetServiceAnnotationsErr error
	ServiceAnnotationsErr    error
	ServiceAnnotationsValue  map[string]string
	IsNotFoundErr            bool
}

func (f *FakeK8sOperations) CloudProviderName() (string, error) {
	return f.CloudProviderNameValue, f.CloudProviderNameErr
}

func (f *FakeK8sOperations) SetServiceAnnotations(namespace, service string, annotations map[string]string) error {
	return f.SetServiceAnnotationsErr
}

func (f *FakeK8sOperations) ServiceAnnotations(namespace, service string) (map[string]string, error) {
	return f.ServiceAnnotationsValue, f.ServiceAnnotationsErr
}

func (f *FakeK8sOperations) IsNotFound(err error) bool {
	return f.IsNotFoundErr
}
