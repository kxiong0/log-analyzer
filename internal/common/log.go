package common

// fluentbit log event
type LogEvent struct {
	Date        float64     `json:"date"`
	Log         string      `json:"log"`
	K8sMetadata K8sMetadata `json:"kubernetes"`
}

type K8sMetadata struct {
	PodName       string                 `json:"pod_name"`
	PodID         string                 `json:"pod_id"`
	Namespace     string                 `json:"namespace_name"`
	Labels        map[string]interface{} `json:"labels"`
	Annotations   map[string]interface{} `json:"annotations"`
	Host          string                 `json:"host"`
	ContainerName string                 `json:"container_name"`
	ContainerImg  string                 `json:"container_image"`
}
