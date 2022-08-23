package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var pageData = []struct {
	name          string
	renderer      string
	template      string
	errorExpected bool
	errorMessage  string
}{
	{"go_page", "go", "home", false, "error rendering go template"},
	{"go_page_no_template", "go", "no-file", true, "no error rendering non-existent go template, when one is expected"},

	{"jet_page", "jet", "home", false, "error rendering jet template"},
	{"jet_page_no_template", "jet", "no-file", true, "no error rendering non-existent jet template, when one is expected"},

	{"invalid_render_engine", "foo", "home", true, "no error rendering with non-existent template engine"},
}

func TestRender_Page(t *testing.T) {

	for _, tt := range pageData {

		r, err := http.NewRequest("GET", "/home", nil)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()

		testRenderer.Renderer = tt.renderer
		testRenderer.RootPath = "./testdata"

		err = testRenderer.Page(w, r, tt.template, nil, nil)

		if tt.errorExpected {
			if err == nil {
				t.Errorf("%s: %s:", tt.name, tt.errorMessage)
			}
		} else {
			if err != nil {
				t.Errorf("%s: %s: %s", tt.name, tt.errorMessage, err.Error())
			}
		}

	}

}

func TestRender_JetPage(t *testing.T) {

	r, err := http.NewRequest("GET", "/home", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.Renderer = "jet"
	testRenderer.RootPath = "./testdata"

	err = testRenderer.Page(w, r, "home", nil, nil)

	if err != nil {
		t.Errorf("error should be nil for a template that exists")
	}

	err = testRenderer.Page(w, r, "invalid-template", nil, nil)
	if err == nil {
		t.Errorf("error should not be nil for invalid template")
	}

}

func TestRender_GoPage(t *testing.T) {

	r, err := http.NewRequest("GET", "/home", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.Renderer = "go"
	testRenderer.RootPath = "./testdata"

	err = testRenderer.Page(w, r, "home", nil, nil)

	if err != nil {
		t.Errorf("error should be nil for a template that exists")
	}

	err = testRenderer.Page(w, r, "invalid-template", nil, nil)
	if err == nil {
		t.Errorf("error should not be nil for invalid template")
	}

}
