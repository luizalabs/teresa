package spec

const (
	defaultPortName = "tcp"
)

type ServicePort struct {
	Name       string
	Port       int
	TargetPort int
}

type Service struct {
	Name         string
	Namespace    string
	Type         string
	Labels       map[string]string
	Ports        []ServicePort
	SourceRanges []string
}

func NewService(namespace, name, sType string, ports []ServicePort, labels map[string]string) *Service {
	return &Service{
		Namespace: namespace,
		Name:      name,
		Type:      sType,
		Ports:     ports,
		Labels:    labels,
	}
}

func NewDefaultService(appName, sType, portName string) *Service {
	ports := []ServicePort{*NewDefaultServicePort(portName)}
	return NewService(
		appName,
		appName,
		sType,
		ports,
		map[string]string{"run": appName},
	)
}

func NewServicePort(name string, port, targetPort int) *ServicePort {
	return &ServicePort{
		Name:       name,
		Port:       port,
		TargetPort: targetPort,
	}
}

func NewDefaultServicePort(name string) *ServicePort {
	if name == "" {
		name = defaultPortName
	}
	return NewServicePort(name, DefaultExternalPort, DefaultPort)
}
