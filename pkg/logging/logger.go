package logging

import (
	"sort"
	"strings"
	"time"

	"github.com/monetr/monetr/pkg/config"
	"github.com/sirupsen/logrus"
)

func NewLoggerWithConfig(configuration config.Logging) *logrus.Entry {
	logger := logrus.New()

	level, err := logrus.ParseLevel(configuration.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	switch strings.ToLower(configuration.Format) {
	default:
		fallthrough
	case "text":
		logger.Formatter = &logrus.TextFormatter{
			ForceColors:               false,
			DisableColors:             false,
			ForceQuote:                true,
			DisableQuote:              false,
			EnvironmentOverrideColors: false,
			DisableTimestamp:          true,
			FullTimestamp:             false,
			TimestampFormat:           "",
			DisableSorting:            false,
			SortingFunc: func(input []string) {
				if len(input) == 0 {
					return
				}
				keys := make([]string, 0, len(input)-1)
				for _, key := range input {
					if key == "msg" {
						continue
					}

					keys = append(keys, key)
				}
				sort.Strings(keys)
				keys = append(keys, "msg")
				copy(input, keys)
			},
			DisableLevelTruncation: false,
			PadLevelText:           true,
			QuoteEmptyFields:       false,
			FieldMap:               nil,
			CallerPrettyfier:       nil,
		}
	case "json":
		formatter := &logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339,
			DisableTimestamp:  false,
			DisableHTMLEscape: false,
			DataKey:           "",
			FieldMap:          logrus.FieldMap{},
			CallerPrettyfier:  nil,
			PrettyPrint:       false,
		}

		// If we are using stackdriver, then use the `message` field name instead. Stackdriver will automatically detect
		// this field in the json payload of a log entry.
		if configuration.StackDriver.Enabled {
			formatter.FieldMap = logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			}
		}
	}

	logger.Formatter = NewContextFormatterWrapper(logger.Formatter)

	if configuration.StackDriver.Enabled {
		formatter, err := NewStackDriverFormatterWrapper(logger.Formatter)
		if err != nil {
			logger.WithError(err).Errorf("failed to create stack driver wrapper")
			return logrus.NewEntry(logger)
		}

		logger.Formatter = formatter
	}

	return logrus.NewEntry(logger)
}

func NewLogger() *logrus.Entry {
	return NewLoggerWithLevel(logrus.InfoLevel.String())
}

func NewLoggerWithLevel(levelString string) *logrus.Entry {
	return NewLoggerWithConfig(config.Logging{
		Level: levelString,
	})
}
