package logger

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	tempFileName := fmt.Sprintf("test-%d.log", os.Getpid())

	Logger := NewLogger(LevelDebug2, tempFileName, true, "test: ", true)
	Logger.SetTimeFromat("")
	Logger.Info("this ", "is ", "info ", "message")
	Logger.Error("this ", "is ", "error ", "message")
	Logger.ErrorL(errors.New("this is errorL message on {f}:{l}"))
	_, filename, line, _ := runtime.Caller(0)
	line -= 1
	exc := fmt.Sprintf(`test: INFO: this is info message
test: EROR: this is error message
test: EROR: %s:%d
test: EROR: C:/Program Files/Go/src/testing/testing.go:1690
test: EROR: C:/Program Files/Go/src/runtime/asm_amd64.s:1700
test: EROR: this is errorL message on logger_test.go:%d
`, filename, line, line)
	exc = strings.ReplaceAll(exc, "\r", "")
	Logger.Close()

	rawData, _ := os.ReadFile(tempFileName)
	data := strings.ReplaceAll(string(rawData), "\r", "")
	if string(data) != exc {
		t.Errorf("expected %s, got %s", exc, data)
	}
	os.Remove(tempFileName)
}

func TestTimeFormat(t *testing.T) {
	tempFileName := fmt.Sprintf("test-%d.log", os.Getpid())
	Logger := NewLogger(LevelDebug2, tempFileName, true, "test: ", true)
	Logger.SetTimeFromat("[15:04:05] ")
	now := time.Now().Format("15:04:05")
	Logger.Info("this ", "is ", "info ", "message")

	rawData, _ := os.ReadFile(tempFileName)
	data := strings.ReplaceAll(string(rawData), "\r", "")
	if !strings.Contains(data, now) {
		t.Errorf("expected %s, got %s", now, data)
	}
	Logger.Close()
	os.Remove(tempFileName)
}

func TestHex(t *testing.T) {
	bytes := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	exp := "[00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f]"
	res := Hex(bytes)
	if res != exp {
		t.Errorf("expected %s, got %s", exp, res)
	}
}

func TestStruct(t *testing.T) {
	type testA struct {
		A int
		B string
	}
	type testB struct {
		A string
		B *string
		С testA
	}
	msg := testB{
		A: "",
		B: nil,
		С: testA{
			A: 1,
			B: "test",
		},
	}
	exp := "{B:<nil> С:{A:1 B:test}}"
	res := Struct(msg)
	if res != exp {
		t.Errorf("expected %s, got %s", exp, res)
	}
}
