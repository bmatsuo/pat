package pat

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
)

func TestPatMatch(t *testing.T) {
	params, ok := (&patHandler{"/", nil}).try("/")
	assert.Equal(t, true, ok)

	params, ok = (&patHandler{"/", nil}).try("/wrong_url")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name", nil}).try("/foo/bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar/")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}, ":id": {"123"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:name", nil}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar", "123"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil}).try("/foo/bar.txt")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name", nil}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil}).try("/foo/val1:val2")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {"val1"}, ":b": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:a.", nil}).try("/foo/.")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}, ":b": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b:c", nil}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}, ":b": {""}, ":c": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/::name", nil}).try("/foo/val1:val2")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":": {"val1"}, ":name": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil}).try("/foo/bar/baz.txt")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/xbar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)
}

func TestPatRoutingHit(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.Query())
		assert.Equal(t, "keith", r.URL.Query().Get(":name"))
	}))

	r, err := http.NewRequest("GET", "/foo/keith?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
}

func TestPatRoutingMethodNotAllowed(t *testing.T) {
	p := New()

	var ok bool
	p.Post("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	p.Put("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	r, err := http.NewRequest("GET", "/foo/keith", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, r)

	assert.T(t, !ok)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	allowed := strings.Split(rr.Header().Get("Allow"), ", ")
	sort.Strings(allowed)
	assert.Equal(t, allowed, []string{"POST", "PUT"})
}

// Check to make sure we don't pollute the Raw Query when we have no parameters
func TestPatNoParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		assert.Equal(t, "", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
}

// Check to make sure we don't pollute the Raw Query when there are parameters but no pattern variables
func TestPatOnlyUserParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		assert.Equal(t, "a=b", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
}

func TestPatImplicitRedirect(t *testing.T) {
	nopHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := func() *http.Request {
		r, err := http.NewRequest("GET", "/foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	p := New()
	p.Get("/foo/", nopHandler)

	res := httptest.NewRecorder()
	p.ServeHTTP(res, req())

	if res.Code != 301 {
		t.Errorf("expected Code 301, was %d", res.Code)
	}

	if loc := res.Header().Get("Location"); loc != "/foo/" {
		t.Errorf("expected %q, got %q", "/foo/", loc)
	}

	p = New()
	p.Get("/foo", nopHandler)
	p.Get("/foo/", nopHandler)

	res = httptest.NewRecorder()
	res.Code = 200
	p.ServeHTTP(res, req())

	if res.Code != 200 {
		t.Errorf("expected Code 200, was %d", res.Code)
	}

	mux := New()
	mux.Add("GET", "/:foo/", nopHandler)
	res = httptest.NewRecorder()
	mux.ServeHTTP(res, req())
	if res.Code != 301 {
		t.Errorf("expected Code 301, was %d", res.Code)
	}
	if loc := res.Header().Get("Location"); loc != "/foo/" {
		t.Errorf("expected %q, got %q", "/foo/", loc)
	}
}

func TestTail(t *testing.T) {
	for i, test := range []struct {
		pat    string
		path   string
		expect string
	}{
		{"/:a/", "/x/y/z", "y/z"},
		{"/:a/", "/x", ""},
		{"/:a/", "/x/", ""},
		{"/:a", "/x/y/z", ""},
		{"/b/:a", "/x/y/z", ""},
		{"/hello/:title/", "/hello/mr/mizerany", "mizerany"},
		{"/:a/", "/x/y/z", "y/z"},
	} {
		tail := Tail(test.pat, test.path)
		if tail != test.expect {
			t.Errorf("failed test %d: Tail(%q, %q) == %q (!= %q)",
				i, test.pat, test.path, tail, test.expect)
		}
	}
}
