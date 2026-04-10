package handlers

import (
	"github.com/samber/mo"

	"identity/domain"
)

type scopeView struct {
	Value       domain.ScopeValue
	Description mo.Option[domain.ScopeDescription]
}

type roleView struct {
	Value       domain.RoleValue
	Description mo.Option[domain.RoleDescription]
}

type appRegistrationView struct {
	Name          domain.AppName
	ClientID      domain.ClientID
	IdentifierURI mo.Option[domain.IdentifierURI]
	RedirectURLs  []domain.RedirectURL
	Scopes        []scopeView
	AppRoles      []roleView
	CsharpConfig  CsharpHostConfig
}

type clientView struct {
	Name     domain.AppName
	ClientID domain.ClientID
}

type groupView struct {
	ID   domain.GroupID
	Name domain.GroupName
}

type userView struct {
	Username    domain.Username
	DisplayName domain.DisplayName
	Email       domain.Email
}

type tenantView struct {
	TenantID         domain.TenantID
	Name             domain.TenantName
	AppRegistrations []appRegistrationView
	Clients          []clientView
	Groups           []groupView
	Users            []userView
}

type indexViewModel struct {
	Tenants []tenantView
}

func optionalIdentifierURI(v domain.IdentifierURI) mo.Option[domain.IdentifierURI] {
	if v.Value() == emptyValue {
		return mo.None[domain.IdentifierURI]()
	}

	return mo.Some(v)
}

func optionalScopeDescription(v domain.ScopeDescription) mo.Option[domain.ScopeDescription] {
	if v.Value() == emptyValue {
		return mo.None[domain.ScopeDescription]()
	}

	return mo.Some(v)
}

func optionalRoleDescription(v domain.RoleDescription) mo.Option[domain.RoleDescription] {
	if v.Value() == emptyValue {
		return mo.None[domain.RoleDescription]()
	}

	return mo.Some(v)
}

func buildScopeViews(scopes []domain.Scope) []scopeView {
	views := make([]scopeView, emptySliceSize, len(scopes))

	for _, s := range scopes {
		view := scopeView{
			Value:       s.Value(),
			Description: optionalScopeDescription(s.Description()),
		}
		views = append(views, view)
	}

	return views
}

func buildRoleViews(roles []domain.Role) []roleView {
	views := make([]roleView, emptySliceSize, len(roles))

	for _, r := range roles {
		view := roleView{
			Value:       r.Value(),
			Description: optionalRoleDescription(r.Description()),
		}
		views = append(views, view)
	}

	return views
}

func buildAppRegistrationViews(tenantID domain.TenantID, apps []domain.AppRegistration) []appRegistrationView {
	views := make([]appRegistrationView, emptySliceSize, len(apps))

	for _, app := range apps {
		view := appRegistrationView{
			Name:          app.Name(),
			ClientID:      app.ClientID(),
			IdentifierURI: optionalIdentifierURI(app.IdentifierURI()),
			RedirectURLs:  app.RedirectURLs(),
			Scopes:        buildScopeViews(app.Scopes()),
			AppRoles:      buildRoleViews(app.AppRoles()),
			CsharpConfig:  NewCsharpHostConfig(tenantID, app.IdentifierURI(), app.ClientID()),
		}
		views = append(views, view)
	}

	return views
}

func buildClientViews(clients []domain.Client) []clientView {
	views := make([]clientView, emptySliceSize, len(clients))

	for _, c := range clients {
		view := clientView{
			Name:     c.Name(),
			ClientID: c.ClientID(),
		}
		views = append(views, view)
	}

	return views
}

func buildGroupViews(groups []domain.Group) []groupView {
	views := make([]groupView, emptySliceSize, len(groups))

	for _, g := range groups {
		view := groupView{
			ID:   g.ID(),
			Name: g.Name(),
		}
		views = append(views, view)
	}

	return views
}

func buildUserViews(users []domain.User) []userView {
	views := make([]userView, emptySliceSize, len(users))

	for _, u := range users {
		display := u.Display()
		view := userView{
			Username:    display.Username,
			DisplayName: display.DisplayName,
			Email:       display.Email,
		}
		views = append(views, view)
	}

	return views
}

func buildIndexViewModel(config domain.Config) indexViewModel {
	tenants := config.Tenants()
	tenantViews := make([]tenantView, emptySliceSize, len(tenants))

	for _, tenant := range tenants {
		view := tenantView{
			TenantID:         tenant.TenantID(),
			Name:             tenant.Name(),
			AppRegistrations: buildAppRegistrationViews(tenant.TenantID(), tenant.AppRegistrations()),
			Clients:          buildClientViews(tenant.Clients()),
			Groups:           buildGroupViews(tenant.Groups()),
			Users:            buildUserViews(tenant.Users()),
		}
		tenantViews = append(tenantViews, view)
	}

	return indexViewModel{Tenants: tenantViews}
}
