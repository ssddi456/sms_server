package logging

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Logger struct {
	level string
}

func NewLogger(level string) *Logger {
	return &Logger{level: strings.ToLower(level)}
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.write("INFO", msg, fields)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.write("WARN", msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.write("ERROR", msg, fields)
}

func (l *Logger) write(level, msg string, fields map[string]interface{}) {
	payload := map[string]interface{}{
		"level": level,
		"msg":   msg,
	}
	for k, v := range fields {
		payload[k] = v
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("%s %s", level, msg)
		return
	}
	log.New(os.Stdout, "", 0).Println(string(b))
}
