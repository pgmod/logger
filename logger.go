package logger

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var cwd string

func init() {
	cwd, _ = os.Getwd()
	cwd = strings.ReplaceAll(cwd, "\\", "/")
}

// Определение уровней логирования с использованием iota
const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
	LevelDebug2
)

// Цветные префиксы для логирования
const (
	MSGprefix   = "\033[34m" + "INFO: " + "\033[0m"
	ERORprefix  = "\033[31m" + "EROR: " + "\033[0m"
	WARNprefix  = "\033[33m" + "WARN: " + "\033[0m"
	DEBGprefix  = "\033[35m" + "DBG1: " + "\033[0m"
	DEBG2prefix = "\033[36m" + "DBG2: " + "\033[0m"
)

type Logger struct {
	level        int
	fileName     string
	needPrefix   bool
	customPrefix string
	file         *os.File
	rewrite      bool
	timeFormat   string
}

// Создание нового логгера
func NewLogger(logLevel int, fileName string, needLevelPrefix bool, loggerPrefix string, rewrite bool) *Logger {
	var f *os.File

	l := Logger{
		level:        logLevel,
		fileName:     fileName,
		needPrefix:   needLevelPrefix,
		customPrefix: loggerPrefix,
		file:         f,
		rewrite:      rewrite,
	}
	l.SetFileName(fileName)
	l.SetTimeFormat("[2006-01-02 15:04:05] ")
	return &l
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) SetFileName(fileName string) {
	l.setFile(fileName, l.rewrite)
}

func (l *Logger) SetNeedPrefix(needPrefix bool) {
	l.needPrefix = needPrefix
}

func (l *Logger) SetCustomPrefix(customPrefix string) {
	l.customPrefix = customPrefix
}

func (l *Logger) SetTimeFormat(timeFormat string) {
	l.timeFormat = timeFormat
}

// Закрытие файла для логгера
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) setFile(fileName string, rewrite bool) {
	l.fileName = fileName
	l.Close()
	if fileName != "" {
		var err error
		if rewrite {
			err = os.Remove(fileName)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Fatalf("error removing file: %v", err)
				}
			}
		}
		l.file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
	}
}

// Вспомогательная функция для логирования
func (l Logger) logMessage(levelPrefix string, levelThreshold int, message ...any) {
	timestamp := time.Now().Format(l.timeFormat)
	log := fmt.Sprintf("%s%s%s%s", timestamp, l.customPrefix, levelPrefix, strings.ReplaceAll(fmt.Sprint(message...), "\n", "\n"+l.customPrefix+levelPrefix))

	_, filename, line, _ := runtime.Caller(2) // столько вложенных функций необходимо до того как программа дойдет до l.logMessage
	filename = strings.ReplaceAll(filename, cwd, "")
	log = strings.ReplaceAll(log, "{f}", strings.TrimLeft(filename, "/"))
	log = strings.ReplaceAll(log, "{l}", fmt.Sprint(line))

	l.writeToFile(log)

	if l.level >= levelThreshold {
		fmt.Println(log)
	}
}

// Функция для логирования стека вызовов и ошибки
func (l Logger) ErrorL(err error) {
	if err == nil {
		return
	}

	var stackTrace strings.Builder
	// stackTrace.WriteString(l.customPrefix + ERORprefix + " " + err.Error() + "\n")

	// Добавляем стек вызовов
	for i := 1; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stackTrace.WriteString(fmt.Sprintf("%s:%d\n", file, line))
	}

	// Добавляем информацию об ошибке
	stackTrace.WriteString(err.Error())

	// Итоговый лог
	logText := stackTrace.String()

	// Записываем и печатаем лог
	if l.level >= LevelError {
		l.logMessage(ERORprefix, LevelError, logText)
	}
}

// Методы для разных уровней логирования
func (l Logger) Error(message ...any) {
	l.logMessage(ERORprefix, LevelError, message...)
}

func (l Logger) Warn(message ...any) {
	l.logMessage(WARNprefix, LevelWarn, message...)
}

func (l Logger) Info(message ...any) {
	l.logMessage(MSGprefix, LevelInfo, message...)
}

func (l Logger) Debug(message ...any) {
	l.logMessage(DEBGprefix, LevelDebug, message...)
}

func (l Logger) Debug2(message ...any) {
	l.logMessage(DEBG2prefix, LevelDebug2, message...)
}

// Запись в файл
func (l Logger) writeToFile(message string) {
	if l.file != nil {
		_, err := l.file.WriteString(clearFromEscapes(message) + "\n")
		if err != nil {
			log.Printf("error writing to file: %v", err)
		}
	}
}

func Struct(msg any) string {
	return filterPrint(fmt.Sprintf("%+v", msg))
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

// Преобразование байтов в шестнадцатеричную строку вида [00 00 00]
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

// Очистка символов escape из строки
func clearFromEscapes(str string) string {
	re := regexp.MustCompile(`\033\[\d+m`)
	return re.ReplaceAllString(str, "")
}
