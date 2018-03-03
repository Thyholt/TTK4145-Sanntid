package assertionCheck

import (
	_ "io"
	_ "io/ioutil"
	"log"
	"os"
	"runtime"
	. "library/colors"
)

var (
	assertErrIntern *log.Logger
)

func init() {

	assertErrIntern = log.New(os.Stderr,
		ColR+"ASSERTION ERROR: "+ColN,
		log.Lmicroseconds)
}

// if condition is true, error message is printed and prosess is killed
func ASSERTION_ERROR(condition bool, err string){
	if condition {
		pc, _, line, _ := runtime.Caller(1)
		assertErrIntern.Printf("%s:%d "+ err, runtime.FuncForPC(pc).Name(), line)
		os.Exit(1)
	}
}

// if condition is true, error message is printed and prosess is killed
func ASSERTION_CHECK(condition bool, err string){
	if condition {
		pc, _, line, _ := runtime.Caller(1)
		assertErrIntern.Printf("%s:%d "+ err, runtime.FuncForPC(pc).Name(), line)
		os.Exit(1)
	}
}
