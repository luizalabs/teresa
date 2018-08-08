package cloudprovider

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/service"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

const (
	awsSSLCertAnnotation  = "service.beta.kubernetes.io/aws-load-balancer-ssl-cert"
	awsSSLPortsAnnotation = "service.beta.kubernetes.io/aws-load-balancer-ssl-ports"
)

type awsOperations struct {
	k8s K8sOperations
}

func (ops *awsOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	hasIngress, err := ops.k8s.HasIngress(appName, appName)
	if err != nil {
		return err
	}
	if hasIngress {
		return ErrNotImplementedOnIngress
	}
	anMap := map[string]string{
		awsSSLCertAnnotation:  cert,
		awsSSLPortsAnnotation: strconv.Itoa(port),
	}
	if err := ops.k8s.SetServiceAnnotations(appName, appName, anMap); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func (ops *awsOperations) SSLInfo(appName string) (*service.SSLInfo, error) {
	var port int
	an, err := ops.k8s.ServiceAnnotations(appName, appName)
	if err != nil {
		if ops.k8s.IsNotFound(err) {
			return nil, ErrServiceNotFound
		}
		return nil, teresa_errors.NewInternalServerError(err)
	}
	sslAn, ok := an[awsSSLPortsAnnotation]
	if ok {
		port, err = strconv.Atoi(sslAn)
		if err != nil {
			return nil, teresa_errors.NewInternalServerError(err)
		}
	}
	info := &service.SSLInfo{
		Cert: an[awsSSLCertAnnotation],
		ServicePort: &spec.ServicePort{
			Port: port,
		},
	}
	return info, nil
}

func (ops *awsOperations) Name() string {
	return "aws"
}
