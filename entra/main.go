package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"html/template"
	"log"
	"net/http"
	"os"
)

var privateKey *rsa.PrivateKey
var configData Config
var loginTemplate *template.Template

func main() {
	const defaultKeyPath = "/certs/identity-signing.key"

	keyPath := os.Getenv("PRIVATE_KEY_PATH")
	if keyPath == "" {
		keyPath = defaultKeyPath
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		log.Fatal("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			log.Fatal("not an RSA private key")
		}
	}

	configData = loadConfig()

	loginTemplate, err = template.ParseFiles("login.html")
	if err != nil {
		log.Fatalf("failed to parse login template: %v", err)
	}

	mux := http.NewServeMux()

	// Discovery and JWKS
	mux.HandleFunc("/_health", health)
	mux.HandleFunc("/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/{tenant}/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/common/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/v2.0/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/{tenant}/v2.0/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/common/v2.0/.well-known/openid-configuration", discovery)
	mux.HandleFunc("/discovery/instance", callHome)
	mux.HandleFunc("/common/discovery/instance", callHome)
	mux.HandleFunc("/{tenant}/discovery/instance", callHome)

	mux.HandleFunc("/discovery/keys", jwks)
	mux.HandleFunc("/common/discovery/keys", jwks)
	mux.HandleFunc("/{tenant}/discovery/keys", jwks)
	mux.HandleFunc("/discovery/v2.0/keys", jwks)
	mux.HandleFunc("/common/discovery/v2.0/keys", jwks)
	mux.HandleFunc("/{tenant}/discovery/v2.0/keys", jwks)

	// Token endpoint
	mux.HandleFunc("/token", token)
	mux.HandleFunc("/{tenant}/oauth2/token", token)
	mux.HandleFunc("/common/oauth2/token", token)
	mux.HandleFunc("/{tenant}/oauth2/v2.0/token", token)
	mux.HandleFunc("/common/oauth2/v2.0/token", token)

	// Interactive flow
	mux.HandleFunc("/authorize", authorize)
	mux.HandleFunc("/{tenant}/oauth2/authorize", authorize)
	mux.HandleFunc("/common/oauth2/authorize", authorize)
	mux.HandleFunc("/{tenant}/oauth2/v2.0/authorize", authorize)
	mux.HandleFunc("/common/oauth2/v2.0/authorize", authorize)
	mux.HandleFunc("/login", login)

	log.Println("Mock Entra ID server starting on :8080")
	if err := http.ListenAndServe(":8080", logger(corsMiddleware(maxBytesMiddleware(mux)))); err != nil {
		log.Fatal(err)
	}
}
