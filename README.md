mgw-core-manager
=======

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=v%2A&label=release)

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=lib%2A&label=latest)

![Image](https://img.shields.io/github/v/tag/SENERGY-Platform/mgw-core-manager?filter=client%2A&label=latest)

Generate Swagger Docs:

    swag init -g routes.go -dir handler/http_hdl/standard,handler/http_hdl/shared --parseDependency --instanceName standard
    swag init -g routes.go -dir handler/http_hdl/restricted,handler/http_hdl/shared --parseDependency --instanceName restricted