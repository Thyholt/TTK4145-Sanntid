package logger

import (
	_ "io"
	_ "io/ioutil"
	"log"
	"os"
	"runtime"
	. "library/colors"
)

var (
	infoIntrn    	*log.Logger
	warningIntrn 	*log.Logger
	errIntrn     	*log.Logger
)

//log.Ldate|
//|log.Lshortfile

func init() {
	infoIntrn = log.New(os.Stdout,
		ColG+"INFO: "+ColN,
		log.Lmicroseconds)

	warningIntrn = log.New(os.Stdout,
		ColY+"WARNING: "+ColN,
		log.Lmicroseconds)

	errIntrn = log.New(os.Stderr,
		ColR+"ERROR: "+ColN,
		log.Lmicroseconds)
}

func INFO(info string) {
	pc, _, line, _ := runtime.Caller(1)
	infoIntrn.Printf("%s:%d "+info, runtime.FuncForPC(pc).Name(), line)
}

func WARNING(warning string) {
	pc, _, line, _ := runtime.Caller(1)
	warningIntrn.Printf("%s:%d "+warning, runtime.FuncForPC(pc).Name(), line)
}

func ERROR(err string) {
	pc, _, line, _ := runtime.Caller(1)
	errIntrn.Printf("%s:%d "+err, runtime.FuncForPC(pc).Name(), line)
	os.Exit(1)
}
