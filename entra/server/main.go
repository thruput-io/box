package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var privateKey *rsa.PrivateKey
var configData Config
var loginTemplate *template.Template
var indexTemplate *template.Template

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
	if err := srv.ListenAndServe(); err != nil {
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

	loginTemplate, err = template.ParseFiles("login.html")
	if err != nil {
		log.Printf("failed to parse login template: %v", err)
	}

	indexTemplate, err = template.ParseFiles("index.html")
	if err != nil {
		log.Printf("failed to parse index template: %v", err)
	}
}
