package model

type Endpoint struct {
	DeploymentID string `json:"d_id"`
	Host         string `json:"host"`
	Port         *int   `json:"port"`
	IntPath      string `json:"int_path"`
	ExtPath      string `json:"ext_path"`
}

type EndpointType = int
