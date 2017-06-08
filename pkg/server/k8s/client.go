package k8s

import (
	"io"

	"github.com/luizalabs/teresa-api/pkg/server/app"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
)

type k8sClient struct {
	kc *kubernetes.Clientset
}

func (k *k8sClient) Create(app *app.App, st st.Storage) error {
	panic("not implemented")
}

func (k *k8sClient) getNamespace(namespace string) (*k8sv1.Namespace, error) {
	ns, err := k.kc.CoreV1().Namespaces().Get(namespace)
	if err != nil {
		return nil, cleanError(err)
	}
	return ns, nil
}

func (k *k8sClient) NamespaceAnnotation(namespace, annotation string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", err
	}

	return ns.Annotations[annotation], nil
}

func (k k8sClient) NamespaceLabel(namespace, label string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", err
	}

	return ns.Labels[label], nil
}

func (k *k8sClient) PodList(namespace string) ([]*app.Pod, error) {
	podList, err := k.kc.CoreV1().Pods(namespace).List(k8sv1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]*app.Pod, 0)
	for _, pod := range podList.Items {
		p := &app.Pod{Name: pod.Name}
		for _, status := range pod.Status.ContainerStatuses {
			if status.State.Waiting != nil {
				p.State = status.State.Waiting.Reason
			} else if status.State.Terminated != nil {
				p.State = status.State.Terminated.Reason
			} else if status.State.Running != nil {
				p.State = string(api.PodRunning)
			}
			if p.State != "" {
				break
			}
		}
		pods = append(pods, p)
	}
	return pods, nil
}

func (k *k8sClient) PodLogs(namespace string, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	req := k.kc.CoreV1().Pods(namespace).GetLogs(
		podName,
		&k8sv1.PodLogOptions{
			Follow:    follow,
			TailLines: &lines,
		},
	)

	return req.Stream()
}

func newInClusterK8sClient() (Client, error) {
	conf, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kc, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	return &k8sClient{kc}, nil
}

func newOutOfClusterK8sClient(conf *Config) (Client, error) {
	k8sConf := &restclient.Config{
		Host:     conf.Host,
		Username: conf.Username,
		Password: conf.Password,
		Insecure: conf.Insecure,
	}
	kc, err := kubernetes.NewForConfig(k8sConf)
	if err != nil {
		return nil, err
	}
	return &k8sClient{kc}, nil
}