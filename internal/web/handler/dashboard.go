package handler

import (
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/internal/web/template"
)

type Dashboard struct {
	tmpl *template.Template
	log  *log.Logger
}

func NewDashboard(tmpl *template.Template, log *log.Logger) *Dashboard {
	return &Dashboard{tmpl: tmpl, log: log}
}

func (d *Dashboard) Template() http.HandlerFunc {
	type data struct {
		FirstName string
		LastName  string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())

		data := data{
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}

		d.tmpl.Execute(w, r, "dashboard", template.ExecuteParams{
			Title:         "Dashboard",
			PageID:        web.PageDashboard,
			NavLinks:      web.NavBarAuthenticated,
			Authenticated: true,
			Data:          data,
		})
	}
}
