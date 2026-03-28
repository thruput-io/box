package main

import (
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed templates/login.html
var loginHTML embed.FS

//go:embed templates/index.html
var indexHTML embed.FS

var (
	privateKey    *rsa.PrivateKey
	configData    Config
	loginTemplate *template.Template
	indexTemplate *template.Template
)

func main() {
	const defaultKeyPath = "/certs/identity-signing.key"

	keyPath := os.Getenv("PRIVATE_KEY_PATH")
	if keyPath == "" {
		keyPath = defaultKeyPath
	}

	loadResources(keyPath)

	mux := setupRouter()

	log.Println("Mock Entra ID server starting on :8080")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      logger(corsMiddleware(maxBytesMiddleware(mux))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func loadResources(keyPath string) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Printf("failed to read key: %v", err)

		return
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		log.Printf("failed to decode PEM block")

		return
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Printf("failed to parse private key: %v", err)

			return
		}
	} else {
		var ok bool

		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			log.Printf("not an RSA private key")

			return
		}
	}

	configData = loadConfig()

	loginTemplate, err = template.ParseFS(loginHTML, "templates/login.html")
	if err != nil {
		log.Printf("failed to parse login template: %v", err)
	}

	indexTemplate, err = template.ParseFS(indexHTML, "templates/index.html")
	if err != nil {
		log.Printf("failed to parse index template: %v", err)
	}
}
