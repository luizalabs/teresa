package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/luizalabs/teresa/tests/k8s"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// used only for debug purpose
func prettyPrintJSON(data interface{}) (string, error) {
	output := &bytes.Buffer{}
	if err := json.NewEncoder(output).Encode(data); err != nil {
		return "", err
	}
	formatted := &bytes.Buffer{}
	if err := json.Indent(formatted, output.Bytes(), "", "  "); err != nil {
		return "", err
	}
	return string(formatted.Bytes()), nil
}

func createBuildForWally() {
	buildPodName := k8s.SlugBuilderPodName("ml-mobile", "123456")
	var pod *api.Pod

	pod = k8s.BuildSlugbuilderPod(
		true, // debug
		buildPodName,
		"default",                    // namespace
		"apps/in/wally-001/app.tgz",  // app tar path
		"apps/out/wally-001/app.tgz", // put path to slug
		"", // buildpacks
	)

	json, err := prettyPrintJSON(pod)
	if err == nil {
		fmt.Printf("Pod spec: %v", json)
	} else {
		fmt.Printf("Error creating json representaion of pod spec: %v", err)
	}

	config := &restclient.Config{
		Host:     "https://k8s-staging.a.luizalabs.com",
		Username: "admin",
		Password: "VOpgP0Ggnty5mLcq",
		Insecure: true,
	}

	c, _ := client.New(config)

	// c.Pods("default")
	// interface que possibilita criar pods dentro de um mesmo namespace
	podsInterface := c.Pods(api.NamespaceDefault)
	newPod, err := podsInterface.Create(pod)
	if err != nil {
		fmt.Printf("creating builder pod (%s)", err)
		panic("Casa caiu")
	}
	fmt.Printf("-> %s\n", newPod.GetName())

	sessionIdleInterval := time.Duration(1) * time.Second
	builderPodTickDuration := time.Duration(1) * time.Second
	builderPodWaitDuration := time.Duration(1) * time.Minute

	if err := k8s.WaitForPod(c, newPod.Namespace, newPod.Name, sessionIdleInterval, builderPodTickDuration, builderPodWaitDuration); err != nil {
		fmt.Printf("watching events for builder pod startup (%s)", err)
	}

	fmt.Println("--> debug... started the build!")

	// log inside container
	if false {
		req := c.Get().Namespace(newPod.Namespace).Name(newPod.Name).Resource("pods").SubResource("log").VersionedParams(
			&api.PodLogOptions{
				Follow: true,
			}, api.ParameterCodec)

		rc, errStream := req.Stream()
		if errStream != nil {
			fmt.Printf("attempting to stream logs (%s)", errStream)
		}
		defer rc.Close()

		size, errCopy := io.Copy(os.Stdout, rc)
		if errCopy != nil {
			fmt.Printf("fetching builder logs (%s)\n", errCopy)
		}

		fmt.Printf("size of streamed logs %v\n", size)
	}

	fmt.Printf(
		"Waiting for the %s/%s pod to end. Checking every %s for %s\n",
		newPod.Namespace,
		newPod.Name,
		builderPodTickDuration,
		builderPodWaitDuration,
	)

	if err := k8s.WaitForPodEnd(c, newPod.Namespace, newPod.Name, builderPodTickDuration, builderPodWaitDuration); err != nil {
		fmt.Printf("error getting builder pod status (%s)", err)
	}

	fmt.Printf("Done\n")
	fmt.Printf("Checking for builder pod exit code\n")

	buildPod, err := c.Pods(newPod.Namespace).Get(newPod.Name)
	if err != nil {
		fmt.Printf("error getting builder pod status (%s)", err)
	}

	for _, containerStatus := range buildPod.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			fmt.Printf("Build pod exited with code %d, stopping build.", state.ExitCode)
		}
	}
	fmt.Printf("Build Complete. Exit code: 0\n")
}

func createDeploymentForWally() {
	d := k8s.BuildSlugRunnerDeployment(
		"wally-runner-eder-1", "default",
		1, 1, 1, "wally-runner-eder-1",
		"apps/out/wally-001/app.tgz/slug.tgz",
		map[string]string{
			"AWS_ACCESS_KEY": "AKIAJVYPW4MGR2DW6QDA",
			"AWS_SECRET_KEY": "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",
		})

	json, err := prettyPrintJSON(d)
	if err == nil {
		fmt.Printf("%v", json)
	} else {
		fmt.Printf("Error creating json representaion of pod spec: %v", err)
	}

	config := &restclient.Config{
		Host:     "https://k8s-staging.a.luizalabs.com",
		Username: "admin",
		Password: "VOpgP0Ggnty5mLcq",
		Insecure: true,
	}

	c, _ := client.New(config)

	deployInterface := c.Deployments(api.NamespaceDefault)

	dd, err := deployInterface.Create(d)

	fmt.Printf(">%v\n", dd)
	fmt.Printf(">%v\n", err)
}

func createServiceForWally() {
	s := k8s.BuildSlugRunnerLBService("wally-service-eder-1", "default", "wally-runner-eder-1")

	json, err := prettyPrintJSON(s)
	if err == nil {
		fmt.Printf("%v", json)
	} else {
		fmt.Printf("Error creating json representaion: %v", err)
	}

	config := &restclient.Config{
		Host:     "https://k8s-staging.a.luizalabs.com",
		Username: "admin",
		Password: "VOpgP0Ggnty5mLcq",
		Insecure: true,
	}

	c, _ := client.New(config)

	si := c.Services(api.NamespaceDefault)

	ns, err := si.Create(s)

	fmt.Printf(">%v\n", ns)
	fmt.Printf(">%v\n", err)
}

func main() {

	// createBuildForWally()

	// createDeploymentForWally()

	createServiceForWally()

}
