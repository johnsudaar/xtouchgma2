package gui

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Logrus hook

func (g *GUI) Levels() []logrus.Level {
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
	// TODO: debug opts

	return levels
}

func (g *GUI) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil
	}
	formater := logrus.TextFormatter{DisableColors: true}
	formatted, err := formater.Format(entry)
	if err != nil {
		return errors.Wrap(err, "fail to format")
	}

	values := strings.Split(string(formatted), "\n")
	g.logChan <- values

	return nil
}

func (g *GUI) startLogs() {
	buffer := []string{}
	for logs := range g.logChan {
		logs := removeEmpty(logs)
		buffer = append(logs, buffer...)
		if len(buffer) > 15 {
			buffer = buffer[:15]
		}
		g.configurationTab.logs.Text = strings.Join(buffer, "\n")
		g.configurationTab.logs.Refresh()
	}
}

func removeEmpty(logs []string) []string {
	res := make([]string, 0)
	for _, s := range logs {
		if s == "" {
			continue
		}
		res = append(res, s)
	}
	return res
}

func (g *GUI) SetStatus(status string) {
	g.configurationTab.SetStatus(status)
}
