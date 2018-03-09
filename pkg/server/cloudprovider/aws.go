package cloudprovider

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

const (
	awsSSLCertAnnotation         = "service.beta.kubernetes.io/aws-load-balancer-ssl-cert"
	awsSSLPortsAnnotation        = "service.beta.kubernetes.io/aws-load-balancer-ssl-ports"
	awsBackendProtocolAnnotation = "service.beta.kubernetes.io/aws-load-balancer-backend-protocol"
	tcpProto                     = "tcp"
)

type awsOperations struct {
	k8s K8sOperations
}

func (ops *awsOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	anMap := map[string]string{
		awsSSLCertAnnotation:         cert,
		awsSSLPortsAnnotation:        strconv.Itoa(port),
		awsBackendProtocolAnnotation: tcpProto,
	}
	if err := ops.k8s.SetServiceAnnotations(appName, appName, anMap); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}
