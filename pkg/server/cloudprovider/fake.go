package cloudprovider

type FakeK8sOperations struct {
	CloudProviderNameErr     error
	CloudProviderNameValue   string
	SetServiceAnnotationsErr error
	SetIngressAnnotationsErr error
	ServiceAnnotationsErr    error
	ServiceAnnotationsValue  map[string]string
	IngressAnnotationsErr    error
	IngressAnnotationsValue  map[string]string
	IsNotFoundErr            bool
	HasIngressValue          bool
	HasIngressErr            error
}

func (f *FakeK8sOperations) CloudProviderName() (string, error) {
	return f.CloudProviderNameValue, f.CloudProviderNameErr
}

func (f *FakeK8sOperations) SetServiceAnnotations(namespace, service string, annotations map[string]string) error {
	return f.SetServiceAnnotationsErr
}

func (f *FakeK8sOperations) SetIngressAnnotations(namespace, service string, annotations map[string]string) error {
	return f.SetIngressAnnotationsErr
}

func (f *FakeK8sOperations) ServiceAnnotations(namespace, service string) (map[string]string, error) {
	return f.ServiceAnnotationsValue, f.ServiceAnnotationsErr
}

func (f *FakeK8sOperations) IngressAnnotations(namespace, ingress string) (map[string]string, error) {
	return f.IngressAnnotationsValue, f.IngressAnnotationsErr
}

func (f *FakeK8sOperations) IsNotFound(err error) bool {
	return f.IsNotFoundErr
}

func (f *FakeK8sOperations) HasIngress(namespace, name string) (bool, error) {
	return f.HasIngressValue, f.HasIngressErr
}

func NewFakeOperations() *FakeK8sOperations {
	return new(FakeK8sOperations)
}
