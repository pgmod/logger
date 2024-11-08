package logger

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
)

const (
	MSGprefix   = "\033[34m" + "INFO: " + "\033[0m"
	ERORprefix  = "\033[31m" + "EROR: " + "\033[0m"
	WARNprefix  = "\033[33m" + "WARN: " + "\033[0m"
	DEBGprefix  = "\033[35m" + "DBG1: " + "\033[0m"
	DEBG2prefix = "\033[36m" + "DBG2: " + "\033[0m"
)

type Logger struct {
	// 0 - ошибки
	// 1 - варны
	// 2 - вывод
	// 3 - отладка
	level        int    `default:"3"`
	fileName     string `default:"proj.log"`
	needPrefix   bool   `default:"true"`
	CustomPrefix string `default:"logger-0: "`
}

func NewLogger(logLevel int, fileName string, needLevelPrefix bool, loggerPrefix string) *Logger {

	return &Logger{
		level:        logLevel,
		fileName:     fileName,
		needPrefix:   needLevelPrefix,
		CustomPrefix: loggerPrefix,
	}
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}
func (l *Logger) SetFileName(fileName string) {
	l.fileName = fileName
}
func (l *Logger) SetNeedPrefix(needPrefix bool) {
	l.needPrefix = needPrefix
}
func ClearFromEscapes(str string) string {
	re := regexp.MustCompile(`\033\[\d+m`)
	return re.ReplaceAllString(str, "")
}
func (l Logger) wf(message ...any) {
	if l.fileName != "" {

		f, err := os.OpenFile(l.fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		_, err = f.WriteString(ClearFromEscapes(fmt.Sprint(message...)) + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}
func filterPrint(str string) string {
	emptyStructsReg := regexp.MustCompile(`(?:{\s*})|(?:\[\s*\])|(\s{2,})`)
	emptyFieldsRegStart := regexp.MustCompile(`\s*(?:\w+\:)\}`)
	emptyFieldsRegEnd := regexp.MustCompile(`{\w+:(?: )`)
	emptyFieldsRegMid := regexp.MustCompile(`\s\w+:\s`)
	need_filter := emptyStructsReg.MatchString(str) || emptyFieldsRegStart.MatchString(str) || emptyFieldsRegEnd.MatchString(str) || emptyFieldsRegMid.MatchString(str)
	for need_filter {
		need_filter = emptyStructsReg.MatchString(str) || emptyFieldsRegStart.MatchString(str) || emptyFieldsRegEnd.MatchString(str) || emptyFieldsRegMid.MatchString(str)
		str = emptyStructsReg.ReplaceAllString(str, "")
		str = emptyFieldsRegStart.ReplaceAllString(str, "}")
		str = emptyFieldsRegEnd.ReplaceAllString(str, "{")
		str = emptyFieldsRegMid.ReplaceAllString(str, " ")
	}
	return str
}
func (l Logger) print(message ...any) {
	fmt.Println(filterPrint(fmt.Sprint(message...)))
}

func (l Logger) Info(message ...any) {
	log := strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.CustomPrefix+MSGprefix)
	l.wf(MSGprefix + log)

	if l.level >= 2 {
		l.print(l.CustomPrefix + MSGprefix + log)
	}
}
func (l *Logger) ErrorL(err error) {
	errText := ""
	for i := 1; i < 10; i++ {
		_, filename, line, _ := runtime.Caller(i)
		if line == 0 {
			break
		}
		errText = errText + "\n" + filename + ":" + fmt.Sprint(line)
	}
	log := strings.ReplaceAll(errText+"\n"+err.Error(), "\n", "\n"+l.CustomPrefix+ERORprefix)
	l.wf(MSGprefix + log)
	if l.level >= 0 {
		if err != nil {

			l.print(log)

		}
	}
}

func Struct(msg any) string {
	return fmt.Sprintf("%+v", msg)
}
func Hex(bytes []byte) string {
	var sb strings.Builder
	sb.WriteString("[")

	for i, b := range bytes {
		if i > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "%02x", b)
	}

	sb.WriteString("]")
	return sb.String()
}
func (l Logger) Debug(message ...any) {
	log := strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.CustomPrefix+DEBGprefix)
	l.wf(DEBGprefix + log)
	if l.level >= 3 {
		if l.needPrefix {

			l.print(l.CustomPrefix + DEBGprefix + log)
		} else {
			l.print(fmt.Sprint(message...))

		}
	}
}
func (l Logger) Debug2(message ...any) {
	log := strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.CustomPrefix+DEBG2prefix)
	l.wf(DEBG2prefix + log)
	if l.level >= 4 {
		if l.needPrefix {
			l.print(l.CustomPrefix + DEBG2prefix + log)
		} else {
			l.print(fmt.Sprint(message...))
		}
	}
}
func (l Logger) Error(message ...any) {
	log := strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.CustomPrefix+ERORprefix)
	l.wf(ERORprefix + log)
	if l.level >= 1 {
		l.print(l.CustomPrefix + ERORprefix + log)
	}
}
func (l Logger) Warn(message ...any) {
	log := strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.CustomPrefix+WARNprefix)
	l.wf(WARNprefix + log)
	if l.level >= 1 {
		l.print(l.CustomPrefix + WARNprefix + log)
	}
}
