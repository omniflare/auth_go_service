package main

import (
	"fmt"

	"github.com/omniflare/auth_go_service/internal/config"
	"github.com/omniflare/auth_go_service/internal/db"
)

func main() {
	envConfig := config.NewEnvConfig()
	db.Init(envConfig)

	fmt.Println("Hello there")
}