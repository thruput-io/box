package app

import (
	"crypto/rsa"
	"html/template"

	"identity/domain"
)

// App holds the immutable dependencies wired at startup.
//
// Note: this is currently a dependency-bag used by the HTTP transport; domain extraction
// and stricter layering will be handled in subsequent refactor steps.
type App struct {
	Config        domain.Config
	Key           *rsa.PrivateKey
	LoginTemplate *template.Template
	IndexTemplate *template.Template
}
