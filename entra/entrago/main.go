package main

import (
	"crypto/rand"
	"crypto/rsa"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	"identity/app"
	"identity/domain"
	identityhttp "identity/http/transport"
)

//go:embed templates/login.html
var loginHTML embed.FS

//go:embed templates/index.html
var indexHTML embed.FS

const rsaKeyBits = 2048

func main() {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		log.Fatalf("failed to generate RSA key: %v", err)
	}

	configPath := "Config.yaml"
	_, statErr := os.Stat(configPath)

	if os.IsNotExist(statErr) {
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

	application := &app.App{
		Config:        config,
		Key:           key,
		LoginTemplate: loginTemplate,
		IndexTemplate: indexTemplate,
	}

	srv := &identityhttp.Server{App: application}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("starting entra mock on %s", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
