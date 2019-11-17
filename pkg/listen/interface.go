package listen

type Interface interface {
	Init(config map[string]string) error
	Subscribe(processor func(body []byte, storageUrl string) error) error
	Close() error
}
