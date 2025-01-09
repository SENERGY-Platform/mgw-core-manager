// Package swagger_docs Code generated by swaggo/swag. DO NOT EDIT
package swagger_docs

import "github.com/swaggo/swag"

const docTemplaterestricted = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache-2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/core-services": {
            "get": {
                "description": "List core services including image and container information.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Core Services"
                ],
                "summary": "List services",
                "responses": {
                    "200": {
                        "description": "services",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "$ref": "#/definitions/model.CoreService"
                            }
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/core-services/{name}": {
            "get": {
                "description": "Get core service including image and container information.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Core Services"
                ],
                "summary": "Get service",
                "parameters": [
                    {
                        "type": "string",
                        "description": "service name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "service",
                        "schema": {
                            "$ref": "#/definitions/model.CoreService"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/core-services/{name}/restart": {
            "patch": {
                "description": "Restart core service container.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Core Services"
                ],
                "summary": "Restart service",
                "parameters": [
                    {
                        "type": "string",
                        "description": "service name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "job ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/endpoints": {
            "get": {
                "description": "Get HTTP endpoint.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "HTTP Endpoints"
                ],
                "summary": "Get endpoint",
                "parameters": [
                    {
                        "type": "string",
                        "description": "endpoint id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "endpoints",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "$ref": "#/definitions/model.Endpoint"
                            }
                        }
                    },
                    "404": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/endpoints-batch}": {
            "delete": {
                "description": "Remove multiple HTTP endpoints.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "HTTP Endpoints"
                ],
                "summary": "Delete endpoints",
                "parameters": [
                    {
                        "type": "string",
                        "description": "comma seperated list of endpoint ids (e.g.: id1,id2,...)",
                        "name": "ids",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "reference value (e.g.: a foreign id)",
                        "name": "ref",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "comma seperated list of labels (e.g.: key1=val1,key2=val2,...)",
                        "name": "labels",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "job ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/endpoints/{id}": {
            "delete": {
                "description": "Remove an HTTP endpoint.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "HTTP Endpoints"
                ],
                "summary": "Delete endpoint",
                "parameters": [
                    {
                        "type": "string",
                        "description": "endpoint id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "job ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/endpoints/{id}/alias": {
            "post": {
                "description": "Create an endpoint alias.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "HTTP Endpoints"
                ],
                "summary": "Create endpoint alias",
                "parameters": [
                    {
                        "type": "string",
                        "description": "endpoint id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "endpoint alias information",
                        "name": "alias",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/model.EndpointAliasReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "job ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/info": {
            "get": {
                "description": "Get basic service and runtime information.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Info"
                ],
                "summary": "Get service info",
                "responses": {
                    "200": {
                        "description": "info",
                        "schema": {
                            "$ref": "#/definitions/lib.SrvInfo"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/jobs": {
            "get": {
                "description": "List all jobs.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jobs"
                ],
                "summary": "List jobs",
                "parameters": [
                    {
                        "type": "string",
                        "description": "status to filter by (pending,running,canceled,completed,error,ok)",
                        "name": "status",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "sort in descending order",
                        "name": "sort_desc",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "list jobs since timestamp",
                        "name": "since",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "list jobs until timestamp",
                        "name": "until",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "jobs",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/lib.Job"
                            }
                        }
                    },
                    "400": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/jobs/{id}": {
            "get": {
                "description": "Get a job.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jobs"
                ],
                "summary": "Get job",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "job",
                        "schema": {
                            "$ref": "#/definitions/lib.Job"
                        }
                    },
                    "404": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/jobs/{id}/cancel": {
            "patch": {
                "description": "Cancels a job.",
                "tags": [
                    "Jobs"
                ],
                "summary": "Cancel job",
                "parameters": [
                    {
                        "type": "string",
                        "description": "job id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "404": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/logs": {
            "get": {
                "description": "List logs of core services not running as containers.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Logs"
                ],
                "summary": "List logs",
                "responses": {
                    "200": {
                        "description": "logs",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "$ref": "#/definitions/model.Log"
                            }
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/logs/{id}": {
            "get": {
                "description": "Get log of a core services.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Logs"
                ],
                "summary": "Get Log",
                "parameters": [
                    {
                        "type": "string",
                        "description": "log id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "maximum number of lines to return",
                        "name": "max_lines",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "log entries",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "lib.Job": {
            "type": "object",
            "properties": {
                "canceled": {
                    "type": "string"
                },
                "completed": {
                    "type": "string"
                },
                "created": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "error": {
                    "$ref": "#/definitions/lib.JobErr"
                },
                "id": {
                    "type": "string"
                },
                "result": {},
                "started": {
                    "type": "string"
                }
            }
        },
        "lib.JobErr": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "lib.MemStats": {
            "type": "object",
            "properties": {
                "alloc": {
                    "type": "integer"
                },
                "alloc_total": {
                    "type": "integer"
                },
                "gc_cycles": {
                    "type": "integer"
                },
                "sys_total": {
                    "type": "integer"
                }
            }
        },
        "lib.SrvInfo": {
            "type": "object",
            "properties": {
                "mem_stats": {
                    "$ref": "#/definitions/lib.MemStats"
                },
                "name": {
                    "type": "string"
                },
                "up_time": {
                    "$ref": "#/definitions/time.Duration"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "model.CoreService": {
            "type": "object",
            "properties": {
                "container": {
                    "$ref": "#/definitions/model.SrvContainer"
                },
                "image": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "model.Endpoint": {
            "type": "object",
            "properties": {
                "ext_path": {
                    "type": "string"
                },
                "host": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "int_path": {
                    "type": "string"
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "location": {
                    "type": "string"
                },
                "parent_id": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                },
                "proxy_conf": {
                    "$ref": "#/definitions/model.ProxyConfig"
                },
                "ref": {
                    "type": "string"
                },
                "string_sub": {
                    "$ref": "#/definitions/model.StringSub"
                },
                "type": {
                    "$ref": "#/definitions/model.EndpointType"
                }
            }
        },
        "model.EndpointAliasReq": {
            "type": "object",
            "properties": {
                "path": {
                    "type": "string"
                }
            }
        },
        "model.EndpointType": {
            "type": "integer",
            "enum": [
                1,
                2,
                3
            ],
            "x-enum-varnames": [
                "StandardEndpoint",
                "AliasEndpoint",
                "DefaultGuiEndpoint"
            ]
        },
        "model.Log": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "service_name": {
                    "type": "string"
                }
            }
        },
        "model.ProxyConfig": {
            "type": "object",
            "properties": {
                "headers": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "read_timeout": {
                    "$ref": "#/definitions/time.Duration"
                },
                "websocket": {
                    "type": "boolean"
                }
            }
        },
        "model.SrvContainer": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "state": {
                    "type": "string"
                }
            }
        },
        "model.StringSub": {
            "type": "object",
            "properties": {
                "filters": {
                    "description": "orgStr:newStr",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "mime_types": {
                    "description": "empty -\u003e all types",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "replace_once": {
                    "description": "false -\u003e replace repeatedly",
                    "type": "boolean"
                }
            }
        },
        "time.Duration": {
            "type": "integer",
            "enum": [
                1,
                1000,
                1000000,
                1000000000
            ],
            "x-enum-varnames": [
                "Nanosecond",
                "Microsecond",
                "Millisecond",
                "Second"
            ]
        }
    }
}`

// SwaggerInforestricted holds exported Swagger Info so clients can modify it
var SwaggerInforestricted = &swag.Spec{
	Version:          "0.8.2",
	Host:             "",
	BasePath:         "/restricted",
	Schemes:          []string{},
	Title:            "Core Manager Public API",
	Description:      "Provides access to public management options for the multi-gateway core.",
	InfoInstanceName: "restricted",
	SwaggerTemplate:  docTemplaterestricted,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInforestricted.InstanceName(), SwaggerInforestricted)
}
