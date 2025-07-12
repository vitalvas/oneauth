package logger

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithoutFile(t *testing.T) {
	log := New("")

	assert.NotNil(t, log)
	assert.IsType(t, &logrus.Logger{}, log)
	assert.True(t, log.ReportCaller)
	assert.IsType(t, &logrus.JSONFormatter{}, log.Formatter)
	assert.Equal(t, os.Stdout, log.Out)
}

func TestNew_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	log := New(logFile)

	assert.NotNil(t, log)
	assert.IsType(t, &logrus.Logger{}, log)
	assert.True(t, log.ReportCaller)

	// Verify file was created
	_, err := os.Stat(logFile)
	assert.NoError(t, err)
}

func TestNew_WithNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "nested", "dir", "test.log")

	log := New(logFile)

	assert.NotNil(t, log)

	// Verify nested directory and file were created
	_, err := os.Stat(logFile)
	assert.NoError(t, err)

	// Verify directory structure
	dirPath := filepath.Dir(logFile)
	info, err := os.Stat(dirPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestJSONFormatterConfiguration(t *testing.T) {
	log := New("")

	formatter, ok := log.Formatter.(*logrus.JSONFormatter)
	require.True(t, ok)

	assert.Equal(t, "@timestamp", formatter.FieldMap[logrus.FieldKeyTime])
	assert.Equal(t, "@level", formatter.FieldMap[logrus.FieldKeyLevel])
	assert.Equal(t, "@message", formatter.FieldMap[logrus.FieldKeyMsg])
	assert.Equal(t, "@caller", formatter.FieldMap[logrus.FieldKeyFunc])
	assert.NotNil(t, formatter.CallerPrettyfier)
}

func TestCallerPrettyfier(t *testing.T) {
	log := New("")
	formatter := log.Formatter.(*logrus.JSONFormatter)

	tests := []struct {
		name         string
		file         string
		function     string
		line         int
		expectedFile string
	}{
		{
			name:         "File with prefix",
			file:         "github.com/vitalvas/oneauth/internal/logger/logger.go",
			function:     "github.com/vitalvas/oneauth/internal/logger.New",
			line:         25,
			expectedFile: "internal/logger/logger.go:25",
		},
		{
			name:         "File without prefix",
			file:         "/usr/local/go/src/runtime/proc.go",
			function:     "runtime.main",
			line:         123,
			expectedFile: "/usr/local/go/src/runtime/proc.go:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &runtime.Frame{
				File:     tt.file,
				Function: tt.function,
				Line:     tt.line,
			}

			function, file := formatter.CallerPrettyfier(frame)

			assert.Equal(t, tt.function, function)
			assert.Equal(t, tt.expectedFile, file)
		})
	}
}

func TestLogLevels(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	log := New(logFile)

	// Test that we can write at different levels without panic
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	// Verify log file has content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}
