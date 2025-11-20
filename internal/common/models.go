package common

type LogEvent struct {
	Date        float64     `json:"date"`
	Log         string      `json:"log"`
	K8sMetadata K8sMetadata `json:"kubernetes"`
}

type K8sMetadata struct {
	PodName       string            `json:"pod_name"`
	Namespace     string            `json:"namespace_name"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	Host          string            `json:"host"`
	ContainerName string            `json:"container_name"`
	ContainerImg  string            `json:"container_image"`
}
