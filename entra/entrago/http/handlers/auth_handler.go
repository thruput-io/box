package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

const (
	formKeyState             = "state"
	emptyType                = ""
	errInvalidCredentialsMsg = "invalid username or password"
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

	redirectURIStr := request.URL.Query().Get("redirect_uri")

	redirectURI, err := domain.NewRedirectURL(redirectURIStr)
	if err != nil {
		return badRequest(err)
	}

	allowedURLs, err := app.FindRedirectURLs(*tenant, clientID)
	if err != nil {
		return badRequest(err)
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return badRequest(err)
	}

	tenantOpt := parseTenantIDOption(tenantIDStr)

	data := loginTemplateData{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		State:        parseOptionalOAuthState(request.URL.Query().Get("state")),
		Scope:        parseScopeValues(request.URL.Query().Get("scope")),
		ResponseType: parseOptionalResponseType(request.URL.Query().Get("response_type")),
		Tenant:       tenantOpt,
		Nonce:        parseOptionalNonce(request.URL.Query().Get("nonce")),
		ResponseMode: parseOptionalResponseMode(request.URL.Query().Get("response_mode")),
		Users:        buildUsersDisplay(*tenant, clientID),
	}

	var buf bytes.Buffer

	err = application.LoginTemplate.Execute(&buf, data)
	if err != nil {
		return internalError("failed to render login page")
	}

	return okHTML(buf.Bytes())
}

type validatedLogin struct {
	tenant      *domain.Tenant
	clientID    domain.ClientID
	redirectURI domain.RedirectURL
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

	redirectURIStr := request.Form.Get("redirect_uri")

	redirectURI, err := domain.NewRedirectURL(redirectURIStr)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeInvalidRequest, "invalid redirect_uri format")
	}

	allowedURLs, err := app.FindRedirectURLs(*tenant, clientID)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeClientNotFound, "client not found")
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return validatedLogin{}, domain.NewError(domain.ErrCodeInvalidRedirectURI, "invalid redirect_uri")
	}

	return validatedLogin{tenant: tenant, clientID: clientID, redirectURI: redirectURI}, nil
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

	user, domErr := authenticateLoginRequest(request, *validated.tenant)
	if domErr != nil {
		return fromDomainError(domErr)
	}

	authCode := app.IssueAuthCode(
		application.Key, *user, validated.clientID, validated.redirectURI,
		parseScopeValues(request.Form.Get("scope")),
		validated.tenant.TenantID(),
		parseOptionalNonce(request.Form.Get("nonce")),
	)

	target := validated.redirectURI.AsURL()

	return buildLoginRedirect(
		&target, authCode,
		parseOptionalOAuthState(request.Form.Get("state")),
		parseOptionalResponseMode(request.Form.Get("response_mode")),
	)
}

func authenticateLoginRequest(request *http.Request, tenant domain.Tenant) (*domain.User, *domain.Error) {
	username, err := domain.NewUsername(request.Form.Get("username"))
	if err != nil {
		return nil, domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg)
	}

	password, err := domain.NewPassword(request.Form.Get("password"))
	if err != nil {
		return nil, domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg)
	}

	user, err := app.AuthenticateUser(tenant, username, password)
	if err != nil {
		return nil, domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg)
	}

	return user, nil
}

func buildLoginRedirect(
	target *url.URL,
	authCode domain.AuthCode,
	state mo.Option[domain.OAuthState],
	responseMode mo.Option[domain.ResponseMode],
) Response {
	values := url.Values{}
	authCode.AddTo(values)

	if s, ok := state.Get(); ok {
		values.Set(formKeyState, s.Value())
	}

	var location string

	isFragment := false
	if rm, ok := responseMode.Get(); ok {
		isFragment = rm.Value() == "fragment"
	}

	if isFragment {
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
	ClientID     domain.ClientID
	RedirectURI  domain.RedirectURL
	State        mo.Option[domain.OAuthState]
	Scope        []domain.ScopeValue
	ResponseType mo.Option[domain.ResponseType]
	Tenant       mo.Option[domain.TenantID]
	Nonce        mo.Option[domain.Nonce]
	ResponseMode mo.Option[domain.ResponseMode]
	Users        []userDisplay
}

// JoinedScope returns the scope values as a space-separated string for template rendering.
func (d loginTemplateData) JoinedScope() string {
	return domain.JoinScopeValues(d.Scope)
}

// StateValue returns the OAuth state string for template rendering, or empty string if absent.
func (d loginTemplateData) StateValue() string {
	if s, ok := d.State.Get(); ok {
		return s.Value()
	}

	return emptyValue
}

// TenantValue returns the tenant ID string for template rendering, or empty string if absent (common).
func (d loginTemplateData) TenantValue() string {
	if t, ok := d.Tenant.Get(); ok {
		return t.Value()
	}

	return emptyValue
}

// NonceValue returns the nonce string for template rendering, or empty string if absent.
func (d loginTemplateData) NonceValue() string {
	if n, ok := d.Nonce.Get(); ok {
		return n.Value()
	}

	return emptyValue
}

// ResponseModeValue returns the response_mode string for template rendering, or empty string if absent.
func (d loginTemplateData) ResponseModeValue() string {
	if rm, ok := d.ResponseMode.Get(); ok {
		return rm.Value()
	}

	return emptyValue
}

// userDisplay is a view model for a user on the login page.
type userDisplay struct {
	Username    domain.Username
	Password    domain.Password
	DisplayName domain.DisplayName
	Roles       []domain.NonEmptyString
}

func buildUsersDisplay(tenant domain.Tenant, clientID domain.ClientID) []userDisplay {
	var activeClient *domain.Client

	client, err := app.FindClient(tenant, clientID)
	if err == nil {
		activeClient = client
	}

	result := make([]userDisplay, emptySliceSize, len(tenant.Users()))

	for _, user := range tenant.Users() {
		display := user.Display()
		result = append(result, userDisplay{
			Username:    display.Username,
			Password:    display.Password,
			DisplayName: display.DisplayName,
			Roles:       resolveDisplayRoles(user, activeClient, tenant),
		})
	}

	return result
}

func resolveDisplayRoles(user domain.User, client *domain.Client, tenant domain.Tenant) []domain.NonEmptyString {
	if client == nil {
		return nil
	}

	appRoles := collectAssignmentRoles(user, *client)
	result := make([]domain.NonEmptyString, emptySliceSize, len(appRoles))

	for appID, roles := range appRoles {
		appName := resolveAppName(tenant, appID)
		roleStrs := make([]string, len(roles))

		for i, r := range roles {
			roleStrs[i] = r.Value()
		}

		display := appName.Value() + ": " + strings.Join(roleStrs, ", ")
		result = append(result, domain.MustNonEmptyString(display))
	}

	return result
}

func collectAssignmentRoles(user domain.User, client domain.Client) map[domain.ClientID][]domain.RoleValue {
	userGroups := make(map[domain.GroupName]bool)

	for _, groupName := range user.Groups() {
		userGroups[groupName] = true
	}

	appRoles := make(map[domain.ClientID][]domain.RoleValue)

	for _, assignment := range client.GroupRoleAssignments() {
		if !userGroups[assignment.GroupName()] {
			continue
		}

		addAssignmentRoles(appRoles, assignment)
	}

	return appRoles
}

func addAssignmentRoles(appRoles map[domain.ClientID][]domain.RoleValue, assignment domain.GroupRoleAssignment) {
	appID := assignment.ApplicationID()
	appRoles[appID] = append(appRoles[appID], assignment.Roles()...)
}

func resolveAppName(tenant domain.Tenant, appID domain.ClientID) domain.AppName {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID() == appID {
			return registration.Name()
		}
	}

	return domain.MustAppName(appID.Value())
}

// parseTenantIDOption wraps the raw tenant ID string in mo.Option[domain.TenantID].
// Returns None for "common" or empty string (resolved to first tenant by FindTenant).
func parseTenantIDOption(raw string) mo.Option[domain.TenantID] {
	if raw == emptyValue || raw == segmentCommon {
		return mo.None[domain.TenantID]()
	}

	tid, err := domain.NewTenantID(raw)
	if err != nil {
		return mo.None[domain.TenantID]()
	}

	return mo.Some(tid)
}
