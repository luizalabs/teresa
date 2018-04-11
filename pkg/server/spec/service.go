package spec

type Service struct {
	Name       string
	Namespace  string
	Type       string
	Labels     map[string]string
	TargetPort int
}

func NewService(namespace, name, sType string, tp int, labels map[string]string) *Service {
	return &Service{
		Namespace:  namespace,
		Name:       name,
		Type:       sType,
		TargetPort: tp,
		Labels:     labels,
	}
}

func NewDefaultService(appName, sType string) *Service {
	return NewService(
		appName,
		appName,
		sType,
		DefaultPort,
		map[string]string{"run": appName},
	)
}
