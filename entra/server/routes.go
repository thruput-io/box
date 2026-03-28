package main

import "net/http"

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Root / Index
	mux.HandleFunc("/", index)

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

	// Config snippets
	mux.HandleFunc("/config/raw", configRaw)
	mux.HandleFunc("/config/{tenantId}/app/{appId}/csharp", configCsharpApp)
	mux.HandleFunc("/config/{tenantId}/app/{appId}/js", configJsApp)
	mux.HandleFunc("/config/{tenantId}/client/{clientId}/csharp", configCsharpClient)

	return mux
}
