package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const name = "tee"

var (
	timestamp bool
	cat       bool
	ignore    bool
)
// handeFlags parses all the flags and sets variables accordingly
func handleFlags() int {

	oflags := os.O_WRONLY | os.O_CREATE

	if cat {
		oflags |= os.O_APPEND
	}

	if ignore {
		signal.Ignore(os.Interrupt)
	}

	return oflags
}

func main() {
	flag.BoolVar(&cat, "a", false, "append the output to the files rather than rewriting them")
	flag.BoolVar(&timestamp, "t", false, "use a timestamp instead of a number")
	flag.BoolVar(&ignore, "i", false, "ignore the SIGINT signal")
	flag.Parse()

	oflags := handleFlags()
	files := make([]*os.File, 0, flag.NArg())
	writers := make([]io.Writer, 0, flag.NArg()+1)

	for _, fname := range flag.Args() {
		fname = getNewName(fname)
		f, err := os.OpenFile(fname, oflags, 0o666)
		if err != nil {
			log.Fatalf("%s: error opening %s: %v", name, fname, err)
		}
		files = append(files, f)
		writers = append(writers, f)
	}
	writers = append(writers, os.Stdout)

	mw := io.MultiWriter(writers...)
	if _, err := io.Copy(mw, os.Stdin); err != nil {
		log.Fatalf("%s: error: %v", name, err)
	}

	for _, f := range files {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: error closing file %q: %v\n", name, f.Name(), err)
		}
	}
}

func getNewName(fname string) string {
	if cat {
		return fname
	}

	if timestamp {
		return fname + "." + TimeStamp()

	}
	cnt := 0
	files, _ := filepath.Glob(fname + "*")
	cnt = len(files)
	return fname + "." + strconv.Itoa(cnt)
}

func TimeStamp() string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return strings.Replace(ts, ":", "", -1) // get rid of offensive colons
}
