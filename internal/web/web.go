package web

// The server side endpoints for Clox.
//
// All handler declarations and redirects should use these constants. If any new endpoints are implemented, append to
// these constant values. This will keep the endpoint values centralized and allow for easy changes.
const (
	URLDashboard      string = "/"
	URLLogin          string = "/login"
	URLGoogleLogin    string = "/login/google"
	URLGoogleCallback string = "/login/google/callback"
	URLRegister       string = "/register"
	URLLogout         string = "/logout"
	URLTokens         string = "/tokens"
	URLTokenResource  string = URLTokens + "/{id}"
)

// The server side app page ID's for Clox. Page IDs refer to the actual page displayed.
//
// Every page that is rendered will have a corresponding page ID. These values will be accessible in the templates.
const (
	PageDashboard string = "dashboard"
	PageLogin     string = "login"
	PageRegister  string = "register"
	PageTokens    string = "tokens"
)

// Link holds a URL and its display value. Link will be injected into templates to navigate the Clox server side app.
type Link struct {
	// The URL to navigate to. This is the href attribute in an html anchor tag.
	URL string

	// The value to be displayed in the template. This is the display value of an anchor tag.
	Value string
}

// NavLink holds a link to be displayed in a navigation bar.
type NavLink struct {
	// The ID of the page the link corresponds to. This is useful when wanting to style an active link.
	PageID string

	// The URL the NavLink will navigate to.
	URL string

	// The value that will be displayed by the template to the user.
	Value string
}

// NavLinkDashboard is a navigation link for the dashboard page.
var NavLinkDashboard = NavLink{
	PageID: PageDashboard,
	URL:    URLDashboard,
	Value:  "Dashboard",
}

// NavLinkTokens is a navigation link for the token page.
var NavLinkTokens = NavLink{
	PageID: PageTokens,
	URL:    URLTokens,
	Value:  "Tokens",
}

// NavLinkLogin is a navigation link for the login page.
var NavLinkLogin = NavLink{
	PageID: PageLogin,
	URL:    URLLogin,
	Value:  "Login",
}

// NavBarAuthenticated is the navigation bar to be displayed once a user is authenticated. Use this navigation bar only
// once a user is authenticated and registered.
var NavBarAuthenticated = []NavLink{NavLinkDashboard, NavLinkTokens}
