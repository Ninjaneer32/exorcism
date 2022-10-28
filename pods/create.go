package pods

import (
	"context"
	"strings"

	"github.com/buzzsurfr/exorcism"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

		// Setup return values and get info from request
		var operations []exorcism.PatchOperation
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &exorcism.Result{Msg: err.Error()}, nil
		}

		daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		log.Infof("%d DaemonSets found", len(daemonsets.Items))
		for _, daemonset := range daemonsets.Items {
			// log.Infof("DaemonSet %s/%s found", daemonset.Namespace, daemonset.Name)
			
			// Ignore DaemonSets for known kubernetes components that do not interface as a sidecar
			ignoredLabelKeys := ["k8s-app"]
			ignoreDaemonSet := false
			for k := range daemonset.ObjectMeta.Labels {
				if contains(ignoredLabelKeys, k) {
					ignoreDaemonSet = true
				}
			}
			if(ignoreDaemonSet) {
				log.Infof("Ignored %s/%s because it's part of the standard kubernetes deployment.", daemonset.Namespace, daemonset.Name)
				continue
			}

			var containers []v1.Container
			containers = append(containers, pod.Spec.Containers...)
			sideCar := daemonset.Spec.Template.Spec.DeepCopy().Containers
			containers = append(containers, sideCar)
			operations = append(operations, exorcism.ReplacePatchOperation("/spec/containers", containers))
			
		}

		Add a simple annotation using `AddPatchOperation`
		metadata := map[string]string{"origin": "fromMutation"}
		operations = append(operations, exorcism.AddPatchOperation("/metadata/annotations", metadata))

		log.Flush()
		return &exorcism.Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}

func contains(set []string, element string) bool {
    for _, v := range set {
        if v == element {
            return true
        }
    }
    return false
}