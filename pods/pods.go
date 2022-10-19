package pods

import (
	"encoding/json"

	"github.com/buzzsurfr/exorcism"

	v1 "k8s.io/api/core/v1"
)

// NewValidationHook creates a new instance of pods validation hook
func NewValidationHook() exorcism.Hook {
	return exorcism.Hook{
		Create: validateCreate(),
	}
}

// NewMutationHook creates a new instance of pods mutation hook
func NewMutationHook() exorcism.Hook {
	return exorcism.Hook{
		Create: mutateCreate(),
	}
}

func parsePod(object []byte) (*v1.Pod, error) {
	var pod v1.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}
