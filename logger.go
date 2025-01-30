package logger

import (
	"fmt"
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
	LevelVerbose
	// Deprecated: use LevelVerbose instead.
	LevelDebug2 = LevelVerbose
)

// Цветные префиксы для логирования
var (
	// Зеленый (Green)
	infoPrefix    = []string{"\033[32m", "I", "\033[0m"}
	errorPrefix   = []string{"\033[31m", "E", "\033[0m"}
	warnPrefix    = []string{"\033[33m", "W", "\033[0m"}
	debugPrefix   = []string{"\033[34m", "D", "\033[0m"}
	verbosePrefix = []string{"\033[93m", "V", "\033[0m"}

	defaultLogger = NewLogger(LevelVerbose, "", true, "", true)
)

type Logger struct {
	level        int
	fileName     string
	needPrefix   bool
	customPrefix string
	file         *os.File
	rewrite      bool
	timeFormat   string
	// Формат логирования
	logFormat string
	// Дополнительное фформатирование для второй и последующей строки
	advLogFormat string
	// Дополнительное фформатирование для последней строки
	endLogFormat string
	// Дополнительное фформатирование для одиночных строк
	singleLogFormat string
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
	l.SetTimeFormat("2006-01-02 15:04:05")
	l.SetLogFormat("[{t}] {p} {s0}{s1}{s1}{s2}: {m}")
	return &l
}

func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}
func (l *Logger) SetLevel(level int) {
	l.level = level
}

func SetFileName(fileName string) {
	defaultLogger.SetFileName(fileName)
}
func (l *Logger) SetFileName(fileName string) {
	l.setFile(fileName, l.rewrite)
}

func SetNeedPrefix(needPrefix bool) {
	defaultLogger.SetNeedPrefix(needPrefix)
}
func (l *Logger) SetNeedPrefix(needPrefix bool) {
	l.needPrefix = needPrefix
}

func SetCustomPrefix(customPrefix string) {
	defaultLogger.SetCustomPrefix(customPrefix)
}
func (l *Logger) SetCustomPrefix(customPrefix string) {
	l.customPrefix = customPrefix
}

func SetTimeFormat(timeFormat string) {
	defaultLogger.SetTimeFormat(timeFormat)
}
func (l *Logger) SetTimeFormat(timeFormat string) {
	l.timeFormat = timeFormat
}

func SetLogFormat(logFormats ...string) {
	defaultLogger.SetLogFormat(logFormats...)
}

// SetLogFormat изменяет формат логирования, используемый логгером.
//
// Первый параметр logFormats[0] задает формат для первого сообщения лога
// в группе связанных лог-сообщений. Если указан только один параметр,
// этот формат используется для всех сообщений в группе.
//
// Второй параметр logFormats[1] задает формат для всех лог-сообщений в группе,
// кроме первого и последнего. Если указаны только два параметра,
// второй параметр используется также для последнего сообщения в группе.
//
// Третий параметр logFormats[2] задает формат для последнего сообщения в группе.
//
// Четвертый параметр logFormats[3] задает формат сообщений, которые не
// входят в группу.
//
// Плейсхолдеры:
//
// {t} - время логирования
//
// {p} - префикс логгера
//
// {s} - уровень логирования
//
// {s0} - цвет уровня логирования
//
// {s1} - символ уровня логирования
//
// {s2} - удаление цвета уровня логирования
//
// {s} - уровень логирования
//
// {m} - сообщение
//
// {f} - имя файла
//
// {l} - номер строки
func (l *Logger) SetLogFormat(logFormats ...string) {
	if len(logFormats) > 0 {
		l.logFormat = logFormats[0]
		l.advLogFormat = logFormats[0]
		l.endLogFormat = logFormats[0]
		l.singleLogFormat = logFormats[0]
	}
	if len(logFormats) > 1 {
		l.advLogFormat = logFormats[1]
		l.endLogFormat = logFormats[1]
	}
	if len(logFormats) > 2 {
		l.endLogFormat = logFormats[2]
	}
	if len(logFormats) > 3 {
		l.singleLogFormat = logFormats[3]
	}
}

func Close() {
	defaultLogger.Close()
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
					fmt.Println("error removing file:", err)
				}
			}
		}
		l.file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("error opening file:", err)
		}
	}
}

// Вспомогательная функция для логирования
func (l Logger) logMessage(levelPrefix []string, levelThreshold int, message ...any) {
	timestamp := time.Now().Format(l.timeFormat)

	sb := strings.Builder{}
	messages := strings.Split(fmt.Sprint(message...), "\n")
	if len(messages) > 0 {
		var log string
		if len(messages) > 1 {
			log = l.logFormat
		} else {
			log = l.singleLogFormat
		}
		log = strings.ReplaceAll(log, "{m}", messages[0])
		sb.WriteString(log)
		sb.WriteString("\n")
		if len(messages) > 1 {
			for _, message := range messages[1 : len(messages)-1] {
				log := l.advLogFormat
				log = strings.ReplaceAll(log, "{m}", message)
				sb.WriteString(log)
				sb.WriteString("\n")
			}
			log := l.endLogFormat
			log = strings.ReplaceAll(log, "{m}", messages[len(messages)-1])
			sb.WriteString(log)
			sb.WriteString("\n")
		}
	}
	log := sb.String()
	_, filename, line, _ := runtime.Caller(2) // столько вложенных функций необходимо до того как программа дойдет до l.logMessage
	filename = strings.ReplaceAll(filename, cwd, "")
	log = strings.ReplaceAll(log, "{f}", strings.TrimLeft(filename, "/"))
	log = strings.ReplaceAll(log, "{l}", fmt.Sprint(line))
	log = strings.ReplaceAll(log, "{t}", timestamp)
	log = strings.ReplaceAll(log, "{p}", l.customPrefix)
	log = strings.ReplaceAll(log, "{s}", levelPrefix[0]+levelPrefix[1]+levelPrefix[2])
	log = strings.ReplaceAll(log, "{s0}", levelPrefix[0])
	log = strings.ReplaceAll(log, "{s1}", levelPrefix[1])
	log = strings.ReplaceAll(log, "{s2}", levelPrefix[2])
	log = log[:len(log)-1]

	l.writeToFile(log)

	if l.level >= levelThreshold {
		fmt.Println(log)
	}
}

func ErrorL(err error)       { defaultLogger.ErrorL(err) }
func Error(message ...any)   { defaultLogger.Error(message...) }
func Warn(message ...any)    { defaultLogger.Warn(message...) }
func Info(message ...any)    { defaultLogger.Info(message...) }
func Debug(message ...any)   { defaultLogger.Debug(message...) }
func Verbose(message ...any) { defaultLogger.Verbose(message...) }

// Функция для логирования стека вызовов и ошибки
func (l Logger) ErrorL(err error) {
	if err == nil {
		return
	}

	var stackTrace strings.Builder
	// stackTrace.WriteString(l.customPrefix + ERORprefix + " " + err.Error() + "\n")

	// Добавляем информацию об ошибке
	stackTrace.WriteString(err.Error() + "\n")

	stackTrace.WriteString(getStackTrace())

	// Итоговый лог
	logText := stackTrace.String()

	// Записываем и печатаем лог
	if l.level >= LevelError {
		l.logMessage(errorPrefix, LevelError, logText)
	}
}

// Методы для разных уровней логирования
func (l Logger) Error(message ...any) {
	l.logMessage(errorPrefix, LevelError, message...)
}

func (l Logger) Warn(message ...any) {
	l.logMessage(warnPrefix, LevelWarn, message...)
}

func (l Logger) Info(message ...any) {
	l.logMessage(infoPrefix, LevelInfo, message...)
}

func (l Logger) Debug(message ...any) {
	l.logMessage(debugPrefix, LevelDebug, message...)
}

// Deprecated: use the Verbose method instead.
func (l Logger) Debug2(message ...any) {
	l.Verbose(message...)
}

func (l Logger) Verbose(message ...any) {
	l.logMessage(verbosePrefix, LevelVerbose, message...)
}

// Запись в файл
func (l Logger) writeToFile(message string) {
	if l.file != nil {
		_, err := l.file.WriteString(clearFromEscapes(message) + "\n")
		if err != nil {
			fmt.Println("error writing to file:", err)
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

// пetStackTrace возвращает строку со стеком вызовов
func getStackTrace() string {
	// Создаем срез для хранения вызовов
	pc := make([]uintptr, 10) // Максимум 10 вызовов в стектрейсе
	n := runtime.Callers(3, pc)

	// Получаем объекты Frame для каждого вызова
	frames := runtime.CallersFrames(pc[:n])

	var stackTrace strings.Builder

	// Проходимся по вызовам и собираем информацию
	for {
		frame, more := frames.Next()
		funcPaths := strings.Split(frame.Function, "/")
		funcName := funcPaths[len(funcPaths)-1]
		stackTrace.WriteString(fmt.Sprintf("at %s in %s:%d", funcName, frame.File, frame.Line))
		if !more {
			break
		} else {
			stackTrace.WriteString("\n")
		}
	}

	return stackTrace.String()
}
