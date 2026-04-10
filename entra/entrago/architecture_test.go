package main

import (
	"testing"

	arch "github.com/matthewmcnew/archtest"
)

const (
	pkgDomain        = "identity/domain"
	pkgApp           = "identity/app"
	pkgHTTP          = "identity/http"
	pkgHTTPHandlers  = "identity/http/handlers"
	pkgHTTPTransport = "identity/http/transport"
	pkgConfig        = "identity/config"
)

// These architectural tests enforce package dependency directions.
// They are black-box, deterministic, and fast.

func TestArchitecture_Domain_DoesNotDependOn_OuterLayers(t *testing.T) {
	t.Parallel()

	pt := arch.Package(t, pkgDomain)
	pt.ShouldNotDependOn(pkgApp)
	pt.ShouldNotDependOn(pkgHTTP)
	pt.ShouldNotDependOn(pkgHTTPHandlers)
	pt.ShouldNotDependOn(pkgHTTPTransport)
	pt.ShouldNotDependOn(pkgConfig)
}

func TestArchitecture_App_DoesNotDependOn_HTTP_Or_Config(t *testing.T) {
	t.Parallel()

	pt := arch.Package(t, pkgApp)
	pt.ShouldNotDependOn(pkgHTTP)
	pt.ShouldNotDependOn(pkgHTTPHandlers)
	pt.ShouldNotDependOn(pkgHTTPTransport)
	pt.ShouldNotDependOn(pkgConfig)
}

func TestArchitecture_HTTP_Handlers_DoesNotDependOn_Transport(t *testing.T) {
	t.Parallel()

	pt := arch.Package(t, pkgHTTPHandlers)
	pt.ShouldNotDependOn(pkgHTTPTransport)
}
