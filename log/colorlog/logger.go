package colorlog

import (
	"io"
	"log"
	"os"
)

type level int

const (
	DEBUG level = iota
	INFO
	WARNING
	ERROR
	FATAL

	flag = log.Ldate | log.Ltime
)

var prefix = map[level]string{
	DEBUG:   "[DEBUG] ",
	INFO:    "[INFO] ",
	WARNING: "[WARN] ",
	ERROR:   "[ERROR] ",
	FATAL:   "[FATAL] ",
}

type Logger map[level]LoggerInterface

func New(out, errOut io.Writer, f Formatter) Logger {
	if out == nil {
		out = os.Stdout
	}

	if errOut == nil {
		errOut = os.Stderr
	}

	if f == nil {
		f = new(ColouredFormatter)
	}

	l := make(map[level]LoggerInterface, 5)
	l[DEBUG] = &Wrapper{lvl: DEBUG, formatter: f, logger: log.New(out, f.GetPrefix(DEBUG)+prefix[DEBUG], flag)}
	l[INFO] = &Wrapper{lvl: INFO, formatter: f, logger: log.New(out, f.GetPrefix(INFO)+prefix[INFO], flag)}
	l[WARNING] = &Wrapper{lvl: INFO, formatter: f, logger: log.New(out, f.GetPrefix(WARNING)+prefix[WARNING], flag)}
	l[ERROR] = &Wrapper{lvl: INFO, formatter: f, logger: log.New(errOut, f.GetPrefix(ERROR)+prefix[ERROR], flag)}
	l[FATAL] = &Wrapper{lvl: INFO, formatter: f, logger: log.New(errOut, f.GetPrefix(FATAL)+prefix[FATAL], flag)}

	return Logger(l)
}

type Wrapper struct {
	lvl       level
	formatter Formatter
	logger    LoggerInterface
}

func (w *Wrapper) Print(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Print(v...)
}

func (w *Wrapper) Printf(format string, v ...interface{}) {
	suffix := w.formatter.GetSuffix(w.lvl)
	v = w.formatter.Format(w.lvl, v...)
	w.logger.Printf("%s"+format+suffix, v...)
}

func (w *Wrapper) Println(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Println(v...)
}

func (w *Wrapper) Fatal(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Fatal(v...)
}

func (w *Wrapper) Fatalf(format string, v ...interface{}) {
	suffix := w.formatter.GetSuffix(w.lvl)
	v = w.formatter.Format(w.lvl, v...)
	w.logger.Fatalf("%s"+format+suffix, v...)
}

func (w *Wrapper) Fatalln(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Fatalln(v...)
}

func (w *Wrapper) Panic(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Fatal(v...)
}

func (w *Wrapper) Panicf(format string, v ...interface{}) {
	suffix := w.formatter.GetSuffix(w.lvl)
	v = w.formatter.Format(w.lvl, v...)
	w.logger.Panicf("%s"+format+suffix, v...)
}

func (w *Wrapper) Panicln(v ...interface{}) {
	v = w.formatter.Format(w.lvl, v...)
	v = append(v, w.formatter.GetSuffix(w.lvl))
	w.logger.Panicln(v...)
}
