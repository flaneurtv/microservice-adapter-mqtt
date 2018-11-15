package process

import (
	"bufio"
	"fmt"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"io"
	"os"
	"os/exec"
	"strings"
)

type service struct {
	name               string
	uuid               string
	host               string
	namespaceListener  string
	namespacePublisher string
	cmdLine            string
	logger             core.Logger
}

func NewService(name, uuid, host, namespaceListener, namespacePublisher, cmdLine string, logger core.Logger) core.Service {
	return &service{
		name:               name,
		uuid:               uuid,
		host:               host,
		namespaceListener:  namespaceListener,
		namespacePublisher: namespacePublisher,
		cmdLine:            cmdLine,
		logger:             logger,
	}
}

func (sp *service) Start(input <-chan string) (output <-chan string, errors <-chan string, err error) {
	parts := strings.Fields(sp.cmdLine)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = []string{fmt.Sprintf("SERVICE_NAME=%s", sp.name),
		fmt.Sprintf("SERVICE_UUID=%s", sp.uuid),
		fmt.Sprintf("SERVICE_HOST=%s", sp.host),
		fmt.Sprintf("NAMESPACE_LISTENER=%s", sp.namespaceListener),
		fmt.Sprintf("NAMESPACE_PUBLISHER=%s", sp.namespacePublisher),
	}
	cmd.Env = append(cmd.Env, os.Environ()...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get stdin: %s", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get stdout: %s", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get stderr: %s", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("can't start command: %s", err)
	}

	if input != nil {
		sp.startWriteTo(stdin, input)
	}
	output = sp.startReadFrom(stdout)
	errors = sp.startReadFrom(stderr)
	return output, errors, nil
}

func (sp *service) startWriteTo(writer io.WriteCloser, input <-chan string) {
	go func() {
		for line := range input {
			_, err := writer.Write([]byte(line + "\n"))
			if err != nil {
				sp.logger.Log(core.LogLevelError, fmt.Sprintf("can't write to std stream: %s", err))
			}
		}
	}()
}

func (sp *service) startReadFrom(reader io.ReadCloser) <-chan string {
	result := make(chan string)
	go func() {
		defer close(result)

		reader := bufio.NewReader(reader)
		for {
			line, err := reader.ReadString('\n')
			if err == nil || (err == io.EOF && line != "") {
				line = strings.TrimSpace(line)
				result <- line
			}
			if err != nil {
				if err != io.EOF {
					sp.logger.Log(core.LogLevelError, fmt.Sprintf("can't read from std stream: %s", err))
				}

				break
			}
		}
	}()
	return result
}
