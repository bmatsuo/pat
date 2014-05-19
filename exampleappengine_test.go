/*
hello.go ported for appengine

this differs from the standard hello.go example in two ways: appengine
already provides an http server for you, obviating the need for the
ListenAndServe call (with associated logging), and the package must not be
called main (appengine reserves package 'main' for the underlying program).
*/
package pat

import (
	"io"
	"net/http"
)

func init() {
	// hello world, the web server
	helloServer := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "hello, "+req.URL.Query().Get(":name")+"!\n")
	}
	m := New()
	m.Get("/hello/:name", http.HandlerFunc(helloServer))

	// Register this pat with the default serve mux so that other packages
	// may also be exported. (i.e. /debug/pprof/*)
	http.Handle("/", m)
}

func ExampleAppEngine() {
	// There's nothing here because the AppEngine runs the show.
}
