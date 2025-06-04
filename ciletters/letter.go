//go:build !solution

package ciletters

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed letter.tmpl
var letterTemplate string

var funcMap = template.FuncMap{
	"split":     strings.Split,
	"join":      strings.Join,
	"reverse":   stringSliceReverse,
	"sliceSafe": sliceSafe,
}

func MakeLetter(n *Notification) (string, error) {
	tmpl, err := template.New("letterTemplate").Funcs(funcMap).Parse(letterTemplate)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	if err := tmpl.Execute(&sb, n); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func stringSliceReverse(s []string) []string {
	r := make([]string, 0, len(s))
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}

func sliceSafe(s []string, start, end int) []string {
	if len(s) < end {
		return s[start:]
	}
	return s[start:end]
}
