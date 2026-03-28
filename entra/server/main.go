package main

import (
	"crypto/rand"
	"crypto/rsa"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	"identity/domain"
	identityhttp "identity/http"
)

//go:embed templates/login.html
var loginHTML embed.FS

//go:embed templates/index.html
var indexHTML embed.FS

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate RSA key: %v", err)
	}

	configPath := "Config.yaml"
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		configPath = "DefaultConfig.yaml"
	}

	config, err := domain.LoadConfig(configPath, "Config.schema.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	loginTemplate, err := template.ParseFS(loginHTML, "templates/login.html")
	if err != nil {
		log.Fatalf("failed to parse login template: %v", err)
	}

	indexTemplate, err := template.ParseFS(indexHTML, "templates/index.html")
	if err != nil {
		log.Fatalf("failed to parse index template: %v", err)
	}

	server := &identityhttp.Server{
		Config:        config,
		Key:           key,
		LoginTemplate: loginTemplate,
		IndexTemplate: indexTemplate,
	}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("starting entra mock on %s", addr)

	if err = http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
