package main

import (
	"context"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
)

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

func readFmtCode(r *http.Request) ([]byte, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	rawSource := r.FormValue("code")
	fmtSource, err := formt([]byte(rawSource))
	if err != nil {
		return nil, err
	}
	return fmtSource, nil
}

func formt(code []byte) ([]byte, error) {
	// TODO: run goimports
	return format.Source(code)
}

func bench(code string) (string, error) {
	arr := regexp.MustCompile("(?m)^func (Benchmark[^\\(]*)").FindAllSubmatchString(code, -1)
	var brr []string
	for _, matches := range arr {
		brr = append(brr, matches[1])
	}
	code += fmt.Sprintf(`
func main() {
	for _,f := range []func(*testing.B){%s} {
		fmt.Println(testing.Benchmark(f))
	}
}
`, strings.Join(brr, ","))
	f, err := ioutil.TempFile(os.TempDir(), "gobenchservice")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	filename := f.Name() + ".go"
	if err := ioutil.WriteFile(filename, code, 0644); err != nil {
		return nil, err
	}
	defer os.Remove(filename)
	return exec.Command("go", "run", filename).CombinedOutput()
}

func main() {
	out, err := bench([]byte(`
package main

import (
	"testing"
	"fmt"
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
