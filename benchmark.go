package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"
)

func handleBench(code []byte) ([]byte, error) {
	mainFn := [][]byte{[]byte("func main() {\noutput := make(map[string]testing.BenchmarkResult)")}
	for _, matches := range regexp.MustCompile("(?m)^func (Benchmark[^\\(]*)").FindAllSubmatch(code, -1) {
		mainFn = append(mainFn, []byte(fmt.Sprintf("output[\"%s\"] = testing.Benchmark(%s)", matches[1], matches[1])))
	}
	mainFn = append(mainFn, []byte("b,_ := json.Marshal(output)\nfmt.Println(string(b))\n}"))
	prog, err := handleFmt(append(code, bytes.Join(mainFn, []byte("\n"))...))
	if err != nil {
		return nil, err
	}
	filename := tempFileName() + ".go"
	defer os.Remove(filename)
	// filename := "a.go"
	if err := ioutil.WriteFile(filename, prog, 0644); err != nil {
		return nil, err
	}
	return exec.Command("go", "run", filename).Output()
}

var (
	fileNamePrefixCount = 1
	fileNameLock        sync.Mutex
)

func tempFileName() string {
	fileNameLock.Lock()
	defer fileNameLock.Unlock()
	fileNamePrefixCount++
	return path.Join(os.TempDir(), fmt.Sprintf("gobenchmark-service-%d.go", fileNamePrefixCount))
}
