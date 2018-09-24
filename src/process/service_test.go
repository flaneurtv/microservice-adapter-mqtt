package process_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/process"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSimpleScript(t *testing.T) {
	scriptFile, _ := ioutil.TempFile("", "script*.go")
	defer os.Remove(scriptFile.Name())

	scriptFile.WriteString(`package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(strings.TrimSpace(line) + "_OUT")
	}
}
`)
	scriptFile.Close()

	sp := process.NewService("process1", "uuid1", "host1", "namespace1", fmt.Sprintf("go run %s", scriptFile.Name()), logger.NewNoOpLogger())

	input := make(chan string)
	output, err := sp.Start(input)

	assert.Nil(t, err)
	assert.NotNil(t, output)

	values := []string{"test1", "test2", "test3", "test4", "test5"}

	go func() {
		for _, value := range values {
			input <- value
		}
	}()

	for _, value := range values {
		outValue := <-output
		assert.Equal(t, value+"_OUT", outValue)
	}
}

func TestHugeValues(t *testing.T) {
	scriptFile, _ := ioutil.TempFile("", "script*.go")
	defer os.Remove(scriptFile.Name())

	scriptFile.WriteString(`package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(strings.TrimSpace(line) + "_OUT")
	}
}
`)
	scriptFile.Close()

	sp := process.NewService("process1", "uuid1", "host1", "namespace1", fmt.Sprintf("go run %s", scriptFile.Name()), logger.NewNoOpLogger())

	input := make(chan string)
	output, err := sp.Start(input)

	assert.Nil(t, err)
	assert.NotNil(t, output)

	n := 10000000
	values := []string{strings.Repeat("test1", n), strings.Repeat("test2", n), strings.Repeat("test3", n)}

	go func() {
		for _, value := range values {
			input <- value
		}
	}()

	for _, value := range values {
		outValue := <-output
		assert.Equal(t, value+"_OUT", outValue)
	}
}

func TestInvalidScript(t *testing.T) {
	sp := process.NewService("process1", "uuid1", "host1", "namespace1", "qwerty123098 run", logger.NewNoOpLogger())

	input := make(chan string)
	_, err := sp.Start(input)

	assert.NotNil(t, err)
}

func TestFailedScript(t *testing.T) {
	scriptFile, _ := ioutil.TempFile("", "script*.go")
	defer os.Remove(scriptFile.Name())

	scriptFile.WriteString(`package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "fail" {
			os.Exit(1)
		}
		fmt.Println(line + "_OUT")
	}
}
`)
	scriptFile.Close()

	sp := process.NewService("process1", "uuid1", "host1", "namespace1", fmt.Sprintf("go run %s", scriptFile.Name()), logger.NewNoOpLogger())

	input := make(chan string)
	output, err := sp.Start(input)

	assert.Nil(t, err)
	assert.NotNil(t, output)

	values := []string{"test1", "test2", "fail", "test4", "test5"}

	go func() {
		for _, value := range values {
			input <- value
		}
	}()

	for i, value := range values {
		outValue, ok := <-output
		if i < 2 {
			assert.True(t, ok)
			assert.Equal(t, value+"_OUT", outValue)
		} else {
			assert.False(t, ok)
		}
	}
}
