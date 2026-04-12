package main

import (
	"crypto/rand"
	"crypto/rsa"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/samber/mo"

	"identity/app"
	"identity/config"
	"identity/domain"
	identityhttp "identity/http/transport"
)

//go:embed templates/*
var templatesFS embed.FS

type runDeps struct {
	keyBits    int
	stat       func(string) (os.FileInfo, error)
	getenv     func(string) string
	logf       func(string, ...any)
	loadConfig func(string) mo.Either[domain.Error, domain.Config]
	listen     func(*http.Server) error
}

func run(deps runDeps) error {
	configPath := resolveConfigPath(deps.stat)

	cfgE := deps.loadConfig(configPath)
	cfg, isRight := cfgE.Right()
	if !isRight {
		domErr, _ := cfgE.Left()
		panic(domErr)
	}

	key, err := rsa.GenerateKey(rand.Reader, deps.keyBits)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	loginTemplate, indexTemplate, err := loadTemplates()
	if err != nil {
		deps.logf("warning: could not load templates: %v", err)
	}

	application := &app.App{
		Config:        &cfg,
		Key:           key,
		LoginTemplate: loginTemplate,
		IndexTemplate: indexTemplate,
	}

	srv := &identityhttp.Server{App: application}
	addr := resolveAddr(deps.getenv)

	if deps.logf != nil {
		deps.logf("starting entra mock on %s", addr)
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: serverReadHeaderTimeout,
	}

	err = deps.listen(httpServer)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func resolveConfigPath(stat func(string) (os.FileInfo, error)) string {
	configPath := "Config.yaml"

	_, err := stat(configPath)
	if err != nil {
		configPath = "DefaultConfig.yaml"
	}

	return configPath
}

func resolveAddr(getenv func(string) string) string {
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func loadTemplates() (login *template.Template, index *template.Template, err error) {
	login, err = template.ParseFS(templatesFS, "templates/login.html")
	if err != nil {
		return nil, nil, fmt.Errorf("parse login template: %w", err)
	}

	index, err = template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		return nil, nil, fmt.Errorf("parse index template: %w", err)
	}

	return login, index, nil
}

func defaultLoadConfig(path string) mo.Either[domain.Error, domain.Config] {
	return config.LoadConfig(path, "")
}

const (
	testRSAKeyBits          = 1024
	serverReadHeaderTimeout = 5 * time.Second
)

func main() {
	const rsaKeyBits = 2048

	err := run(runDeps{
		keyBits:    rsaKeyBits,
		stat:       os.Stat,
		getenv:     os.Getenv,
		logf:       log.Printf,
		loadConfig: defaultLoadConfig,
		listen: func(server *http.Server) error {
			return server.ListenAndServe()
		},
	})
	if err != nil {
		const empty = ""

		msg := strings.ReplaceAll(err.Error(), "\n", empty)
		msg = strings.ReplaceAll(msg, "\r", empty)
		log.Fatalf("%s", msg)
	}
}
