package main

import (
	"pluto-backend/internal/manifest/bootstrap"
)

func main() {
	if err := bootstrap.RunManifestService(); err != nil {
		panic(err)
	}
}
