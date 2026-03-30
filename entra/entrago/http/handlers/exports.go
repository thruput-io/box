package handlers

import (
	nethttp "net/http"

	"identity/app"
)

// Exported entrypoints used by the router (http package).

// Health handles the health endpoint.
func Health(request *nethttp.Request, application *app.App) Response {
	return healthHandler(request, application)
}

// Index handles the root endpoint (and delegates to config/test-tokens subroutes).
func Index(request *nethttp.Request, application *app.App) Response {
	return indexHandler(request, application)
}

// Authorize handles OAuth2 authorize requests.
func Authorize(request *nethttp.Request, application *app.App) Response {
	return authorizeHandler(request, application)
}

// Login handles the login form submission.
func Login(request *nethttp.Request, application *app.App) Response {
	return loginHandler(request, application)
}

// Token handles OAuth2 token requests.
func Token(request *nethttp.Request, application *app.App) Response {
	return tokenHandler(request, application)
}

// Discovery handles OIDC discovery requests.
func Discovery(request *nethttp.Request, application *app.App) Response {
	return discoveryHandler(request, application)
}

// JWKS handles JWKS key discovery.
func JWKS(request *nethttp.Request, application *app.App) Response {
	return jwksHandler(request, application)
}

// CallHome handles Entra "call home" discovery requests.
func CallHome(request *nethttp.Request, application *app.App) Response {
	return callHomeHandler(request, application)
}
