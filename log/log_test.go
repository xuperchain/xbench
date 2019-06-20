package log

import (
	"testing"
)

func TestLogNormal(t *testing.T) {
	DEBUG.Println("test debug println!")
	INFO.Printf("test %s!", "info printlf")
	WARNING.Printf("test %s!", "warning printf")
	WARN.Print("test warn print!")
	ERROR.Print("test error print!")
	// ERROR.Fatal("test error fatal!")
	// ERROR.Panicf("test %s!", "error panic")
}

func TestLogPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("log error panic failed!")
		} else {
			INFO.Print("test error panic successful!")
		}
	}()

	ERROR.Panicf("test %s!", "error panic")
}

// TODO
func TestLogFatal(t *testing.T) {

}
