package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func Init(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

var already_initialized bool = false

func DefaultLoggerInit() {
	if already_initialized == false {
		Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
		already_initialized = true
	}
}

func TrackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	Info.Printf("%s %d\n", name, int(elapsed/1000000))
}

func JSONPrintDS(input_data interface{}) {
	data, err := json.Marshal(input_data)
	if err == nil {
		fmt.Println(string(data[:]))
	} else {
		log.Printf("got err %v while marshaling", err)
	}
}
