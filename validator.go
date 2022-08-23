package ugo

import (
	"github.com/asaskevich/govalidator"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

func (u *Ugo) Validator(data url.Values) *Validation {

	return &Validation{
		Errors: make(map[string]string),
		Data:   data,
	}
}

func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validation) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

func (v *Validation) Has(field string, r *http.Request) bool {
	return r.Form.Get(field) != ""
}

func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, f := range fields {
		val := r.Form.Get(f)
		if strings.TrimSpace(val) == "" {
			v.AddError(f, "this field cannot be blank")
		}
	}
}

func (v *Validation) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "invalid email address")
	}
}

func (v *Validation) IsInt(field, value string) {
	if _, err := strconv.Atoi(value); err != nil {
		v.AddError(field, "This field must be an integer")
	}
}

func (v *Validation) IsFloat(field, value string) {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		v.AddError(field, "This field must be a floating point number")
	}
}

func (v *Validation) IsDateISO(field, value string) {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		v.AddError(field, "must be a date in the form of YYYY-MM-DD")
	}
}

func (v *Validation) NoSpaces(field, value string) {
	if !govalidator.HasWhitespace(value) {
		v.AddError(field, "spaces are not allowed in this field")
	}
}
