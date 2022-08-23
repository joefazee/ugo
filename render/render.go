package render

import (
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"html/template"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
	JetViews   *jet.Set
	Session    *scs.SessionManager
}

type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
	Error           string
	Flash           string
}

func (r *Render) defaultData(td *TemplateData, rq *http.Request) *TemplateData {
	td.Secure = r.Secure
	td.ServerName = r.ServerName
	td.CSRFToken = nosurf.Token(rq)
	td.Port = r.Port
	if r.Session != nil && r.Session.Exists(rq.Context(), "userID") {
		td.IsAuthenticated = true
	}

	td.Error = r.Session.PopString(rq.Context(), "error")
	td.Flash = r.Session.PopString(rq.Context(), "flash")
	return td
}

// Page renders a template based on the selected template engine. go or jet
func (r *Render) Page(w http.ResponseWriter, rq *http.Request, view string, variables, data interface{}) error {

	switch strings.ToLower(r.Renderer) {
	case "go":
		return r.GoPage(w, rq, view, data)
	case "jet":
		return r.JetPage(w, rq, view, variables, data)
	default:

	}
	return errors.New("invalid template engine specified")
}

// GoPage renders a template using the standard Go template engine
func (r *Render) GoPage(w http.ResponseWriter, rq *http.Request, view string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.html", r.RootPath, view))
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	err = tmpl.Execute(w, &td)
	if err != nil {
		return err
	}
	return nil
}

// JetPage render`s a template using the Jet template engine
func (r *Render) JetPage(w http.ResponseWriter, rq *http.Request, view string, variables, data interface{}) error {

	var vars jet.VarMap
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = r.defaultData(td, rq)

	t, err := r.JetViews.GetTemplate(fmt.Sprintf("%s.page.jet", view))
	if err != nil {
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		return err
	}

	return nil
}
