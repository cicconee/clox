package template

import (
	"html/template"
	"strings"
)

// funcMap holds all the functions that will be available to the templates. Every template will have access to these
// functions.
var funcMap template.FuncMap = template.FuncMap{
	"formatName": func(name string) string {
		if len(name) == 0 {
			return name
		}

		firstLetter := strings.ToUpper(name[0:1])
		remaining := name[1:]

		return firstLetter + remaining
	},
}
