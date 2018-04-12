package spec

type ServicePort struct {
	Name       string
	Port       int
	TargetPort int
}

type Service struct {
	Name      string
	Namespace string
	Type      string
	Labels    map[string]string
	Ports     []ServicePort
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

func NewDefaultService(appName, sType string) *Service {
	ports := []ServicePort{{TargetPort: DefaultPort}}
	return NewService(
		appName,
		appName,
		sType,
		ports,
		map[string]string{"run": appName},
	)
}
