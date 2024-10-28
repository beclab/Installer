package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

type CommandExecute interface {
	Execute() (string, error)
}

type CommandExecutor struct {
	cmd         []string
	exitCode    int
	printOutput bool
	printLine   bool
}

func NewCommandExecutor(args []string, printOutput, printLine bool) *CommandExecutor {
	return &CommandExecutor{
		cmd:         args,
		printOutput: printOutput,
		printLine:   printLine,
	}
}

func (command *CommandExecutor) getCmd() string {
	return strings.Join(command.cmd, " ")
}

func (command *CommandExecutor) Execute() (string, error) {
	args := append([]string{"/C"}, command.cmd...)
	c := exec.Command("cmd", args...)

	out, err := c.StdoutPipe()
	if err != nil {
		return "", err
	}

	c.Stderr = c.Stdout

	if err := c.Start(); err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
		return "", err
	}

	var outputBuffer bytes.Buffer
	r := bufio.NewReader(out)

	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line) + "\r"
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("[exec] read error: %s", err)
			}
			if command.printLine && line != "" {
				fmt.Println(line)
			}
			outputBuffer.WriteString(line)
			break
		}
		if command.printLine && line != "" {
			fmt.Println(line)
		}
		outputBuffer.WriteString(line)
	}

	err = c.Wait()
	if err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
	}
	res := outputBuffer.String()

	if command.printOutput {
		fmt.Printf("[exec] CMD: %s, OUTPUT: \n%s\n", c.String(), res)
	}
	logger.Infof("[exec] CMD: %s, OUTPUT: %s", c.String(), res)
	return res, errors.Wrapf(err, "Failed to exec command: %s \n%s", command.getCmd(), res)
}

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}
