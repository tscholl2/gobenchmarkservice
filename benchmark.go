package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"
	"testing"
	"text/template"
)

const programTemplate = `{{printf "%s" .Code}}
func main() {
	output := make(map[string]testing.BenchmarkResult)
	{{range $key,$value := .Benchmarks}}
	output["{{printf "%s" $value}}"] = testing.Benchmark({{printf "%s" $value}})
	{{end}}
	json.NewEncoder(os.Stdout).Encode(output)
}`

var benchmarkLock sync.Mutex

// handleBench runs all the benchmarks in the given code and returns the result
// as a map of the form benchmark name -> result.
// The input code should define usual benchmark functions of the form
//
//    func BenchmarkFoo(b *testing.B) {
//    	for i := 0; i < b.N; i++ {
//    		foo()
//    	}
//    }
//
// The resulting map will have contain a key "BenchmarkFoo" with the results
// from running this benchmark.
// Note: handleBench should only be run one at a time to avoid benchmarks
// interferring with eachother. Otherwise, usual benchmark precautions should be taken
// (i.e. don't benchmark if the machine is super busy).
func handleBench(code []byte) (map[string]testing.BenchmarkResult, error) {
	benchmarkLock.Lock()
	defer benchmarkLock.Unlock()
	t := template.Must(template.New("program").Parse(programTemplate))
	var benchmarks [][]byte
	for _, matches := range regexp.MustCompile("(?m)^func (Benchmark[^\\(]*)").FindAllSubmatch(code, -1) {
		benchmarks = append(benchmarks, matches[1])
	}
	var buf bytes.Buffer
	t.Execute(&buf, map[string]interface{}{"Code": code, "Benchmarks": benchmarks})
	prog, err := handleFmt(buf.Bytes())
	if err != nil {
		return nil, err
	}
	filename := path.Join(os.TempDir(), "gobenchmarkservice.go")
	defer os.Remove(filename)
	if err := ioutil.WriteFile(filename, prog, 0644); err != nil {
		return nil, err
	}
	out, err := exec.Command("go", "run", filename).Output()
	m := make(map[string]testing.BenchmarkResult)
	json.Unmarshal(out, &m)
	return m, err
}
