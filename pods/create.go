package pods

import (
	"context"
	"strings"

	"github.com/buzzsurfr/exorcism"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	log "k8s.io/klog/v2"
)

func validateCreate() exorcism.AdmitFunc {
	return func(r *v1beta1.AdmissionRequest) (*exorcism.Result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &exorcism.Result{Msg: err.Error()}, nil
		}

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				return &exorcism.Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
			}
		}

		return &exorcism.Result{Allowed: true}, nil
	}
}

func mutateCreate() exorcism.AdmitFunc {
	return func(r *v1beta1.AdmissionRequest) (*exorcism.Result, error) {
		// Get the list of DaemonSets

		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		log.Infof("%d DaemonSets found", daemonsets.Size())
		for _, daemonset := range daemonsets.Items {
			// fmt.Printf("%s\t%s\n", daemonset.Namespace, daemonset.Name)
			log.Infof("DaemonSet %s/%s found", daemonset.Namespace, daemonset.Name)
		}

		var operations []exorcism.PatchOperation
		// pod, err := parsePod(r.Object.Raw)
		// if err != nil {
		// 	return &exorcism.Result{Msg: err.Error()}, nil
		// }

		// Very simple logic to inject a new "sidecar" container.
		// if pod.Namespace == "special" {
		// 	var containers []v1.Container
		// 	containers = append(containers, pod.Spec.Containers...)
		// 	sideC := v1.Container{
		// 		Name:    "test-sidecar",
		// 		Image:   "busybox:stable",
		// 		Command: []string{"sh", "-c", "while true; do echo 'I am a container injected by mutating webhook'; sleep 2; done"},
		// 	}
		// 	containers = append(containers, sideC)
		// 	operations = append(operations, exorcism.ReplacePatchOperation("/spec/containers", containers))
		// }

		// Add a simple annotation using `AddPatchOperation`
		// metadata := map[string]string{"origin": "fromMutation"}
		// operations = append(operations, exorcism.AddPatchOperation("/metadata/annotations", metadata))

		klog.Flush()
		return &exorcism.Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}
