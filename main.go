package main

import (
	"context"

	"github.com/harness-community/drone-read-trusted/plugin"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(formatter))

	var args plugin.Args
	if err := envconfig.Process("", &args); err != nil {
		logrus.Fatalln(err)
	}

	if err := plugin.Exec(context.Background(), args); err != nil {
		logrus.Fatalln(err)
	}
}

// A simple formatter that prints the message without timestamp.
type formatter struct{}

func (*formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}
