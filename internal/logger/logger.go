package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/internal/tools"
)

const prefixCallerTrim = "github.com/vitalvas/oneauth/"

func New(fileName string) *logrus.Logger {
	log := logrus.New()
	log.SetReportCaller(true)
	log.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
			logrus.FieldKeyFunc:  "@caller",
		},
		CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
			row := strings.Split(frame.File, prefixCallerTrim)

			file := frame.File
			if len(row) == 2 {
				file = row[1]
			}

			// function, file
			return frame.Function, fmt.Sprintf("%s:%d", file, frame.Line)
		},
	})

	if fileName != "" {
		fileDir := filepath.Dir(fileName)
		if err := tools.MkDir(fileDir, 0700); err != nil {
			log.Fatal(err)
		}

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}

	return log
}
