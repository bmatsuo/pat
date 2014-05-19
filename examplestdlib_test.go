package pat

import (
	"io"
	"log"
	"net/http"
)

func ExampleStdlib() {
	helloServer := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "hello, "+req.URL.Query().Get(":name")+"!\n")
	}
	m := New()
	m.Get("/hello/:name", http.HandlerFunc(helloServer))

	// Register this pat with the default serve mux so that other packages
	// may also be exported. (i.e. /debug/pprof/*)
	http.Handle("/", m)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
