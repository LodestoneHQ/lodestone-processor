package listen

type Interface interface {
	Init(config map[string]string) error
	Subscribe() error
	Close() error
}
