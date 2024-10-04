module github.com/SENERGY-Platform/mgw-core-manager

go 1.22

require (
	github.com/SENERGY-Platform/gin-middleware v0.4.3
	github.com/SENERGY-Platform/go-cc-job-handler v0.1.2
	github.com/SENERGY-Platform/go-service-base/config-hdl v0.1.1
	github.com/SENERGY-Platform/go-service-base/context-hdl v0.0.3
	github.com/SENERGY-Platform/go-service-base/job-hdl v1.1.1
	github.com/SENERGY-Platform/go-service-base/job-hdl/lib v0.1.0
	github.com/SENERGY-Platform/go-service-base/logger v0.2.0
	github.com/SENERGY-Platform/go-service-base/srv-info-hdl v0.0.3
	github.com/SENERGY-Platform/go-service-base/srv-info-hdl/lib v0.0.2
	github.com/SENERGY-Platform/go-service-base/util v1.1.0
	github.com/SENERGY-Platform/go-service-base/watchdog v0.4.2
	github.com/SENERGY-Platform/mgw-container-engine-wrapper/client v0.15.1
	github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib v0.16.0
	github.com/SENERGY-Platform/mgw-core-manager/lib v0.0.0-00010101000000-000000000000
	github.com/gin-contrib/requestid v1.0.2
	github.com/gin-gonic/gin v1.10.0
	github.com/tufanbarisyildirim/gonginx v0.0.0-20231222202608-ba16e88a9436
	github.com/y-du/go-env-loader v0.5.2
	github.com/y-du/go-log-level v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/SENERGY-Platform/go-base-http-client v0.0.2 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/SENERGY-Platform/mgw-core-manager/lib => ./lib
