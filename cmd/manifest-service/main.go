package main

import (
	"pluto-backend/internal/manifest/bootstrap"
)

func main() {
	if err := bootstrap.RunManifestService("configs/manifest.yaml"); err != nil {
		panic(err)
	}
}
