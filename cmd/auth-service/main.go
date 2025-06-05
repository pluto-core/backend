package main

import "pluto-backend/internal/auth/bootstrap"

func main() {
	if err := bootstrap.RunAuthService(); err != nil {
		panic(err)
	}
}
