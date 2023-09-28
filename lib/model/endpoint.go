package model

type EndpointType = int

type Endpoint struct {
	ID      string       `json:"id"`
	Type    EndpointType `json:"type"`
	Ref     string       `json:"ref"`
	Host    string       `json:"host"`
	Port    *int         `json:"port"`
	IntPath string       `json:"int_path"`
	ExtPath string       `json:"ext_path"`
}
