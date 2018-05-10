package main

import (
	"fmt"
	"net/http"
)

func readFmtCode(r *http.Request) ([]byte, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	rawSource := r.FormValue("code")
	fmtSource, err := handleFmt([]byte(rawSource))
	if err != nil {
		return nil, err
	}
	return fmtSource, nil
}

func main() {
	out, err := handleBench([]byte(`
package main

import (
	"testing"
)

func add(a,b int) int {
	return a+b
}

func sub(a,b int) int {
	return a-b
}

func BenchmarkAdd(b *testing.B) {
	for i:=0;i<b.N;i++ {
		add(3,7)
	}
}

func BenchmarkSub(b *testing.B) {
	for i:=0;i<b.N;i++ {
		sub(3,7)
	}
}
`))
	fmt.Printf("%s\n%+v\n", out, err)
}

/*
func server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/fmt", func(w http.ResponseWriter, r *http.Request) {
		code, err := readFmtCode(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := w.Write(code); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/bench", func(w http.ResponseWriter, r *http.Request) {
		code, err := readFmtCode(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		out, err := bench(code)
		if _, err := w.Write(out); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	s := http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	s.Shutdown(context.TODO())
}
*/
