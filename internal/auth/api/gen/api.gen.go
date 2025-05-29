package gen

//go:generate oapi-codegen -package gen -generate types -o types.gen.go ../../../../api/openapi/auth.yaml

//go:generate oapi-codegen -package gen -generate chi-server -o server.gen.go ../../../../api/openapi/auth.yaml

//go:generate oapi-codegen -package gen -generate spec -o spec.gen.go ../../../../api/openapi/auth.yaml
