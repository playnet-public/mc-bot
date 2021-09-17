package kubernetes

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// StatefulSetRestarter allows to restart by patching a statefulset
type StatefulSetRestarter struct {
	Namespace    string
	Name         string
	ClientSet    *kubernetes.Clientset
	FieldManager string
}

// Restart patches the statefulset to cause a restart
func (r StatefulSetRestarter) Restart() error {
	// .Spec.Template.ObjectMeta.Annotations["restarter.play-net.org/restartedAt"]
	patch := []byte(fmt.Sprintf(`[{"op": "add", "path": "/spec/template/metadata/annotations/restarter.play-net.org~1restartedAt", "value": %s}]`, time.Now().UTC().Format(time.RFC3339)))
	if _, err := r.ClientSet.AppsV1().StatefulSets(r.Namespace).Patch(context.TODO(), r.Name, types.JSONPatchType, patch, v1.PatchOptions{
		FieldManager: r.FieldManager,
	}); err != nil {
		return err
	}
	return nil
}

// PodRestarter allows to restart by deleting pods matching a certain label
type PodRestarter struct {
	Namespace  string
	LabelKey   string
	LabelValue string
	ClientSet  *kubernetes.Clientset
}

// Restart the app by deleting its pod
func (r PodRestarter) Restart() error {
	req, err := labels.NewRequirement(r.LabelKey, selection.Equals, []string{r.LabelValue})
	if err != nil {
		return err
	}

	return r.ClientSet.CoreV1().Pods(r.Namespace).DeleteCollection(context.TODO(), v1.DeleteOptions{}, v1.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req).String(),
	})
}
