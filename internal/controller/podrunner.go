package controller

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodLabel map[string]string

type PodRunner interface {
	FindDjangoPod(ctx context.Context, namespace string) (*corev1.Pod, error)
	ExecInPod(ctx context.Context, pod *corev1.Pod, command []string) error
}

type DjangoPodRunner struct {
	Client    client.Client
	RESTCfg   *rest.Config
	Clientset *kubernetes.Clientset
	Label     PodLabel
}

// Returns the first Pod that matches the label
func (r DjangoPodRunner) FindDjangoPod(ctx context.Context, ns string) (*corev1.Pod, error) {
	// Find the Django pod in this namespace
	podList := &corev1.PodList{}
	sel := labels.SelectorFromSet(labels.Set(r.Label))
	if err := r.Client.List(ctx, podList, &client.ListOptions{
		Namespace:     ns,
		LabelSelector: sel,
	}); err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return nil, nil
	}
	pod := podList.Items[0]

	return &pod, nil
}

// execInPod runs the given command in the first container of the pod
func (r DjangoPodRunner) ExecInPod(ctx context.Context, pod *corev1.Pod, command []string) error {
	req := r.Clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   command,
			Container: pod.Spec.Containers[0].Name,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(r.RESTCfg, "POST", req.URL())
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
