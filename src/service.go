package core

type Service interface {
	Start(input <-chan string) (output <-chan string, err error)
}
