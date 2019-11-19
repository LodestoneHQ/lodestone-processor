package listen

type Interface interface {
	Init(config map[string]string) error
	Subscribe(processor func(body []byte) error) error
	Close() error
}
