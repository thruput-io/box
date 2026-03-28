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

	"identity/app"
	"identity/domain"
	identityhttp "identity/http/transport"
)

//go:embed templates/login.html
var loginHTML embed.FS

//go:embed templates/index.html
var indexHTML embed.FS

const rsaKeyBits = 2048

const (
	configPath        = "Config.yaml"
	defaultConfigPath = "DefaultConfig.yaml"
	schemaPath        = "Config.schema.json"
	defaultAddr       = ":8080"
)

type runDeps struct {
	keyBits int
	stat    func(string) (os.FileInfo, error)
	getenv  func(string) string
	logf    func(string, ...any)

	loadConfig func(string) (domain.Config, error)
	listen     func(*http.Server) error
}

func defaultLoadConfig(path string) (domain.Config, error) {
	config, err := domain.LoadConfig(path, schemaPath)
	if err != nil {
		return domain.Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	return config, nil
}

func resolveConfigPath(stat func(string) (os.FileInfo, error)) string {
	_, err := stat(configPath)
	if os.IsNotExist(err) {
		return defaultConfigPath
	}

	return configPath
}

func resolveAddr(getenv func(string) string) string {
	if port := getenv("PORT"); port != "" {
		return ":" + port
	}

	return defaultAddr
}

func loadTemplates() (loginTemplate, indexTemplate *template.Template, err error) {
	loginTemplate, err = template.ParseFS(loginHTML, "templates/login.html")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse login template: %w", err)
	}

	indexTemplate, err = template.ParseFS(indexHTML, "templates/index.html")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse index template: %w", err)
	}

	return loginTemplate, indexTemplate, nil
}

func run(deps runDeps) error {
	key, err := rsa.GenerateKey(rand.Reader, deps.keyBits)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	path := resolveConfigPath(deps.stat)

	config, err := deps.loadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	loginTemplate, indexTemplate, err := loadTemplates()
	if err != nil {
		return err
	}

	application := &app.App{
		Config:        config,
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
		Addr:    addr,
		Handler: srv.Handler(),
	}

	err = deps.listen(httpServer)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func main() {
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
		log.Fatalf("%v", err)
	}
}
