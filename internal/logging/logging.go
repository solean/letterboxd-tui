package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const (
	appDir       = "letterboxd-tui"
	errorLogFile = "errors.log"
)

func ErrorLogPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDir, errorLogFile), nil
}

func LogError(context string, err error) {
	if err == nil {
		return
	}

	path, pathErr := ErrorLogPath()
	if pathErr != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	ctx := strings.TrimSpace(context)
	if ctx == "" {
		ctx = "error"
	}
	timestamp := time.Now().Format(time.RFC3339)
	message := formatError(err)
	fmt.Fprintf(file, "%s [%s] %s\n", timestamp, ctx, message)
}

func formatError(err error) string {
	message := strings.TrimSpace(safeErrorString(err))
	if message == "" {
		return "unknown error"
	}
	message = strings.ReplaceAll(message, "\r\n", "\n")
	lines := strings.Split(message, "\n")
	if len(lines) == 1 {
		return lines[0]
	}
	for i := 1; i < len(lines); i++ {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}

func safeErrorString(err error) (message string) {
	if err == nil {
		return "unknown error"
	}
	value := reflect.ValueOf(err)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return "unknown error"
		}
	}
	defer func() {
		if recover() != nil {
			message = "unknown error"
		}
	}()
	return err.Error()
}
