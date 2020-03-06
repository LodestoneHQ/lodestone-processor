package listen

import "github.com/sirupsen/logrus"

type Interface interface {
	Init(logger *logrus.Entry, config map[string]string) error
	Subscribe(processor func(body []byte) error) error
	Close() error
}
