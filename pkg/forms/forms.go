package forms

import (
	"net/url"
)

type errors map[string][]string

func (e errors) Get(field string) string {
	if len(e[field]) == 0 {
		return ""
	}
	return e[field][0]
}

func (e errors) Add(field string, err string) {
	e[field] = append(e[field], err)
}

type Form struct {
	url.Values
	Errors errors
}

func New(form url.Values) *Form {
	return &Form{Values: form, Errors: make(errors)}
}

func (f *Form) Required(name string) {
	if len(f.Get(name)) == 0 {
		f.Errors.Add(name, "required field is empty")
	}
}

func (f *Form) Valid() bool {
	for name := range f.Errors {
		if f.Errors.Get(name) != "" {
			return false
		}
	}
	return true
}
