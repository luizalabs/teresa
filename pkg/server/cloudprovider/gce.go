package cloudprovider

type gceOperations struct {
	k8s K8sOperations
}

func (ops *gceOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	return ErrNotImplemented
}
