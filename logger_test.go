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

	Logger := NewLogger(LevelDebug2, tempFileName, true, "test", true)
	Logger.SetLogFormat("{p}: {s0}{s1}{s1}{s2}: {m}")

	Logger.Info("this ", "is ", "info ", "message")
	Logger.Warn("this ", "is ", "wаrning ", "message")
	Logger.Debug("this ", "is ", "dеbug ", "message")
	Logger.Verbose("this ", "is ", "verbose ", "message")
	Logger.Error("this ", "is ", "error ", "message")
	Logger.ErrorL(errors.New("this is errorL message on {f}:{l}"))
	_, filename, line, _ := runtime.Caller(0)
	line -= 1
	exc := fmt.Sprintf(`test: II: this is info message
test: WW: this is wаrning message
test: DD: this is dеbug message
test: VV: this is verbose message
test: EE: this is error message
test: EE: this is errorL message on logger_test.go:%d
test: EE: at logger.TestFile in %s:%d
test: EE: at testing.tRunner in C:/Program Files/Go/src/testing/testing.go:1690
test: EE: at runtime.goexit in C:/Program Files/Go/src/runtime/asm_amd64.s:1700`+"\n", line, filename, line)
	exc = strings.ReplaceAll(exc, "\r", "")
	Logger.Close()

	rawData, _ := os.ReadFile(tempFileName)
	data := strings.ReplaceAll(string(rawData), "\r", "")
	if data != exc {
		fmt.Println("got\n", Hex(rawData), "\nexpected\n", Hex([]byte(exc)))
		t.Errorf("expected %s, got %s", exc, data)
	}
	os.Remove(tempFileName)
}

func TestFormat(t *testing.T) {
	tempFileName := fmt.Sprintf("test-%d.log", os.Getpid())
	Logger := NewLogger(LevelVerbose, tempFileName, true, "test", true)
	Logger.Debug("this is default debug message")
	Logger.Debug("this is another debug message")
	fmt.Println("")
	Logger.SetTimeFormat("15:04:05")
	now := time.Now().Format("15:04:05")

	Logger.SetLogFormat("{s0}{s1}┐{s2}[{t} in {p}]:\n {s0}└{m}{s2}")
	Logger.Info("this ", "is ", "info ", "message")
	fmt.Println("")
	Logger.SetLogFormat("{t}/{p}/{s1}/{s0}{m}{s2}")
	Logger.Warn(`this
is
big
warning
message`)

	fmt.Println("")
	Logger.SetLogFormat(
		"{s}[{t}]{s0}{p} ┬─{m}{s2}",
		"               {s0} ├─{m}{s2}",
		"               {s0} └─{m}{s2}",

		"{s}[{t}]{s0}{p} ──{m}{s2}",
	)
	Logger.Info("1\n2\n3\n4")
	Logger.Warn("1\n2\n3")
	Logger.Error("a\nb\nc")
	Logger.Debug("1\n2")
	Logger.Verbose("1")

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

func TestDefaultLogger(t *testing.T) {
	Warn("test", "lol")
}
