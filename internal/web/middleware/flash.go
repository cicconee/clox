package middleware

import (
	"net/http"

	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/flash"
)

type Flash struct {
	cookies *cookie.Manager
}

func NewFlash(cookies *cookie.Manager) *Flash {
	return &Flash{cookies: cookies}
}

// Extract will consume all flashes. The message and error flash is retrieved from cookies and injected into the http
// request context. These cookies are cleared after being consumed.
//
// Wrap all http handlers that need to display flash alerts.
func (f *Flash) Extract(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := f.extract(r, cookie.FlashMessage)
		err := f.extract(r, cookie.FlashError)

		// Due to the nature of posting a form with javascript the X-Requested-With header must be checked. If not the
		// cookie will be lost.
		//
		// With a typical HTML form submission the browser makes a synchronous request to the server. If this
		// request responds with a 302 the browser automatically follows it. The cookies are handled automatically.
		// The cookies will be set and included immediately in the next request.
		//
		// With a AJAX (Fetch API or XMLHttpRequest) the form data is sent asynchronously. If the server responds
		// with a 302, the AJAX request does not automatically follow it. The javascript code must manually handle
		// the redirect. Due to the manual redirect, it results in an additional request. By checking that request
		// was made by an AJAX call, the cookie remains available on the next page render.
		//
		// This check results with the cookie not being cleared prematurely.
		if r.Header.Get("X-Requested-With") != "FetchAPI" {
			f.cookies.Clear(w, cookie.FlashMessage)
			f.cookies.Clear(w, cookie.FlashError)
		}

		r = r.WithContext(flash.SetMessageContext(r.Context(), msg))
		r = r.WithContext(flash.SetErrorContext(r.Context(), err))

		next(w, r)
	}
}

// extract gets a flash from cookies. The http.ErrNoCookie error is ignored.
func (f *Flash) extract(r *http.Request, key string) string {
	var flash string

	flashCookie, err := f.cookies.Get(r, key)
	if err != nil {
		return ""
	}
	if flashCookie != nil {
		flash = flashCookie.Value
	}

	return flash
}
