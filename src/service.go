package core

type Service interface {
	Start(input <-chan string) (output <-chan string, errors <-chan string, err error)
}
