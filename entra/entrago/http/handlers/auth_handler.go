package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"identity/app"
	"identity/domain"
)

const (
	formKeyState   = "state"
	emptySliceSize = 0
	emptyType      = ""
)

func authorizeHandler(request *http.Request, application *app.App) Response {
	err := validateParamLengths(request.URL.Query())
	if err != nil {
		return badRequest(err)
	}

	clientIDStr := request.URL.Query().Get("client_id")
	tenantIDStr := extractTenantID(request)

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return badRequest(err)
	}

	clientID, err := domain.NewClientID(clientIDStr)
	if err != nil {
		return badRequest(err)
	}

	redirectURI := request.URL.Query().Get("redirect_uri")

	allowedURLs, err := app.FindRedirectURLs(tenant, clientID)
	if err != nil {
		return badRequest(err)
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return badRequest(err)
	}

	data := loginTemplateData{
		ClientID:     clientIDStr,
		RedirectURI:  redirectURI,
		State:        request.URL.Query().Get("state"),
		Scope:        request.URL.Query().Get("scope"),
		ResponseType: request.URL.Query().Get("response_type"),
		Tenant:       tenantIDStr,
		Nonce:        request.URL.Query().Get("nonce"),
		ResponseMode: request.URL.Query().Get("response_mode"),
		Users:        buildUsersDisplay(tenant, clientID),
	}

	var buf bytes.Buffer

	err = application.LoginTemplate.Execute(&buf, data)
	if err != nil {
		return internalError("failed to render login page")
	}

	return okHTML(buf.Bytes())
}

type validatedLogin struct {
	tenant      domain.Tenant
	clientID    domain.ClientID
	redirectURI string
	tenantIDStr string
}

func validateLoginRequest(request *http.Request, application *app.App) (validatedLogin, *domain.Error) {
	tenantIDStr := request.Form.Get("tenant")

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeTenantNotFound, "tenant not found")
	}

	clientID, err := domain.NewClientID(request.Form.Get("client_id"))
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeInvalidRequest, err.Error())
	}

	redirectURI := request.Form.Get("redirect_uri")

	allowedURLs, err := app.FindRedirectURLs(tenant, clientID)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeClientNotFound, "client not found")
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeInvalidRedirectURI, "invalid redirect_uri")
	}

	return validatedLogin{tenant: tenant, clientID: clientID, redirectURI: redirectURI, tenantIDStr: tenantIDStr}, nil
}

func loginHandler(request *http.Request, application *app.App) Response {
	if request.Method != http.MethodPost {
		return methodNotAllowed()
	}

	request.Body = http.MaxBytesReader(nil, request.Body, app.MaxBodyBytes)

	err := request.ParseForm()
	if err != nil {
		return badRequest(err)
	}

	paramErr := validateParamLengths(request.Form)
	if paramErr != nil {
		return badRequest(paramErr)
	}

	validated, domErr := validateLoginRequest(request, application)
	if domErr != nil {
		return fromDomainError(domErr)
	}

	user, err := app.AuthenticateUser(validated.tenant, request.Form.Get("username"), request.Form.Get("password"))
	if err != nil {
		return fromDomainError(domain.NewError(domain.ErrCodeInvalidCredentials, "invalid username or password"))
	}

	authCode := app.IssueAuthCode(
		application.Key, user, validated.clientID, validated.redirectURI,
		request.Form.Get("scope"), validated.tenantIDStr, request.Form.Get("nonce"),
	)

	target, err := url.Parse(validated.redirectURI)
	if err != nil {
		return badRequest(err)
	}

	return buildLoginRedirect(target, authCode, request.Form.Get("state"), request.Form.Get("response_mode"))
}

func buildLoginRedirect(target *url.URL, authCode, state, responseMode string) Response {
	values := url.Values{}
	values.Set("code", authCode)

	if state != emptyType {
		values.Set(formKeyState, state)
	}

	var location string

	if responseMode == "fragment" {
		target.RawQuery = ""
		location = target.String() + "#" + values.Encode()
	} else {
		query := target.Query()
		for key, vals := range values {
			query.Set(key, vals[0])
		}

		target.RawQuery = query.Encode()
		location = target.String()
	}

	return Response{
		Status:      http.StatusFound,
		ContentType: emptyType,
		Body:        nil,
		Headers:     map[string]string{"Location": location},
	}
}

// loginTemplateData is the data passed to the login template.
type loginTemplateData struct {
	ClientID     string
	RedirectURI  string
	State        string
	Scope        string
	ResponseType string
	Tenant       string
	Nonce        string
	ResponseMode string
	Users        []userDisplay
}

// userDisplay is a view model for a user on the login page.
type userDisplay struct {
	Username    string
	Password    string
	DisplayName string
	Roles       []string
}

func buildUsersDisplay(tenant domain.Tenant, clientID domain.ClientID) []userDisplay {
	var activeClientPtr *domain.Client

	client, err := app.FindClient(tenant, clientID)
	if err == nil {
		activeClientPtr = &client
	}

	result := make([]userDisplay, emptySliceSize, len(tenant.Users()))

	for _, user := range tenant.Users() {
		result = append(result, userDisplay{
			Username:    user.Username().String(),
			Password:    user.Password().String(),
			DisplayName: user.DisplayName().String(),
			Roles:       resolveDisplayRoles(user, activeClientPtr, tenant),
		})
	}

	return result
}

func resolveDisplayRoles(user domain.User, client *domain.Client, tenant domain.Tenant) []string {
	if client == nil {
		return nil
	}

	appRoles := collectAssignmentRoles(user, *client)
	result := make([]string, emptySliceSize, len(appRoles))

	for appIDStr, roles := range appRoles {
		appName := resolveAppName(tenant, appIDStr)
		result = append(result, appName+": "+strings.Join(roles, ", "))
	}

	return result
}

func collectAssignmentRoles(user domain.User, client domain.Client) map[string][]string {
	userGroups := make(map[string]bool)

	for _, groupName := range user.Groups() {
		userGroups[groupName.String()] = true
	}

	appRoles := make(map[string][]string)

	for _, assignment := range client.GroupRoleAssignments() {
		if !userGroups[assignment.GroupName().String()] {
			continue
		}

		appIDStr := assignment.ApplicationID().String()

		for _, roleValue := range assignment.Roles() {
			appRoles[appIDStr] = append(appRoles[appIDStr], roleValue.String())
		}
	}

	return appRoles
}

func resolveAppName(tenant domain.Tenant, appIDStr string) string {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID().String() == appIDStr {
			return registration.Name().String()
		}
	}

	return appIDStr
}
