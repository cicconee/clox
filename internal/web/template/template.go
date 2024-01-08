package template

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/flash"
)

// Template holds the parsed templates and is responsible for executing them. Before executing a template the Parse
// method must be called.
type Template struct {
	Name   string
	Path   string
	Logger *log.Logger
	tmpl   *template.Template
}

// New creates a new Template that will wrap the parsed templates (tmpl).
func New(name string, path string, logger *log.Logger) *Template {
	return &Template{Name: name, Path: path, Logger: logger}
}

// Parse will parse every template (.gohtml) in this template Path. Every template that will be explicitly executed
// should include {{define "template-name"}}. The handlers could then execute these templates by
// specifying the "template-name".
func (t *Template) Parse() error {
	templates := template.New(t.Name).Funcs(funcMap)

	err := filepath.Walk(t.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a ".gohtml" extension.
		if !info.IsDir() && strings.HasSuffix(path, ".gohtml") {
			// Read the template file.
			tmplBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Parse the template and add it to the templates set.
			_, err = templates.Parse(string(tmplBytes))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	t.tmpl = templates

	return nil
}

// ExecuteParams is the data that can be set for the templates. This is the data that changes depending on what page
// is being executed.
type ExecuteParams struct {
	// Title is the page title. It is the value of <title> in the base layout.
	Title string

	// PageID is the ID of the page being executed.
	PageID string

	// NavLinks is the slice of web.NavLink that should be displayed in the navigation bar.
	NavLinks []web.NavLink

	// Authenticated is used by the base layout to render template data that should only be visible if the page requires
	// authentication or not. Each template that requires the user to be authenticated should set this value to true so
	// that all the appropriate html is rendered.
	Authenticated bool

	// Data is the values that will be injected into the page template. Data is not accessible in the base layout.
	Data any
}

// page is the data that is accessible in the page template. This data is not accessible in the base layout template.
//
// Use the Data field to pass the page dependent values to the template. All other fields will hold the same values
// on each page execution.
type page struct {
	// Data holds the page dependent data.
	Data any
}

// base is the data that is accessible the base layout template.
type base struct {
	// Title is the page title. It is the value of <title> in the <head>.
	Title string

	// Content holds the page template as HTML. Content is written to the {{.Content}} directive within the base layout.
	Content template.HTML

	// NavBar is a slice of NavLink's to be displayed in a navigation bar.
	NavBar []web.NavLink

	// PageID is the page that is being displayed within the base layout.
	PageID string

	// Authenticated renders certain data differently based on if the page requires authentication or not.
	Authenticated bool

	// LinkLogout is the logout url and value to be rendered when rendering a page that requires authentication.
	LinkLogout web.Link

	FlashMessage string
	FlashError   string
}

// Execute will execute the specified template defined with tmplName. ExecuteParams is used to inject data into the
// templates.
//
// All templates to be executed will be injected into the base layout template (base.gohtml). The base layout has a
// {{.Content}} directive within the <main> html tag. This is where the templates specified by tmplName will be written.
func (t *Template) Execute(w http.ResponseWriter, r *http.Request, tmplName string, p ExecuteParams) {
	// Build content to be injected into the base layout. This is the current page that is being rendered.
	// Cache it in a buffer to then be injected into the base layout.
	var content bytes.Buffer
	err := t.tmpl.ExecuteTemplate(&content, tmplName, page{
		Data: p.Data,
	})
	if err != nil {
		t.Logger.Printf("[ERROR] [%s %s] Executing template [name: %s]: %v\n", r.Method, r.URL.Path, tmplName, err)
	}

	// Write the base layout with the injected the page template.
	// Page template is injected via Content field.
	err = t.tmpl.ExecuteTemplate(w, "base", base{
		Title:         p.Title,
		Content:       template.HTML(content.String()),
		NavBar:        p.NavLinks,
		PageID:        p.PageID,
		Authenticated: p.Authenticated,
		LinkLogout:    web.Link{URL: web.URLLogout, Value: "Logout"},
		FlashMessage:  flash.GetMessageContext(r.Context()),
		FlashError:    flash.GetErrorContext(r.Context()),
	})
	if err != nil {
		t.Logger.Printf("[ERROR] [%s %s] Executing base template [name: %s]: %v\n", r.Method, r.URL.Path, "base", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
