package helm

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getSecret(client kubernetes.Interface, namespace, name, key string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	secret, err := client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return "", fmt.Errorf("error getting secret %s from namespace %s: %s", name, namespace, err)
	}
	val, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("couldn't find key %s in secret", key)
	}
	return string(val), nil
}
