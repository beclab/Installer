package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
	Run() (string, error)
	Exec() (string, error)
}

type CommandExecutor struct {
	name        string
	prefix      string
	cmd         []string
	exitCode    int
	printOutput bool
	printLine   bool
}

type PowerShellCommandExecutor struct {
	Commands    []string
	PrintOutput bool
	PrintLine   bool
}

func (p *PowerShellCommandExecutor) Run() (string, error) {
	var cmd = &CommandExecutor{
		name:        "powershell",
		prefix:      "-Command",
		cmd:         p.Commands,
		printOutput: p.PrintOutput,
		printLine:   p.PrintLine,
	}

	return cmd.run()
}

type DefaultCommandExecutor struct {
	Commands    []string
	PrintOutput bool
	PrintLine   bool
}

func (d *DefaultCommandExecutor) Run() (string, error) {
	var cmd = &CommandExecutor{
		name:        "cmd",
		prefix:      "/C",
		cmd:         d.Commands,
		printOutput: d.PrintOutput,
		printLine:   d.PrintLine,
	}

	return cmd.run()
}

func (d *DefaultCommandExecutor) Exec() (string, error) {
	var cmd = &CommandExecutor{
		name:        "cmd",
		prefix:      "/C",
		cmd:         d.Commands,
		printOutput: d.PrintOutput,
		printLine:   d.PrintLine,
	}

	return cmd.exec()
}

func NewCommandExecutor(name, prefix string, args []string, printOutput, printLine bool) *CommandExecutor {
	return &CommandExecutor{
		name:        name,
		prefix:      prefix,
		cmd:         args,
		printOutput: printOutput,
		printLine:   printLine,
	}
}

func (command *CommandExecutor) getCmd() string {
	return strings.Join(command.cmd, " ")
}

func (command *CommandExecutor) run() (string, error) {
	args := append([]string{command.prefix}, command.cmd...)
	c := exec.Command(command.name, args...)

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
	logger.Debugf("[exec] CMD: %s, OUTPUT: %s", c.String(), res)
	return res, errors.Wrapf(err, "Failed to exec command: %s \n%s", command.getCmd(), res)
}

func (command *CommandExecutor) exec() (string, error) {
	args := append([]string{command.prefix}, command.cmd...)
	c := exec.Command(command.name, args...)

	out, err := c.StdoutPipe()
	if err != nil {
		return "", err
	}

	_, pipeWriter, err := os.Pipe()
	defer pipeWriter.Close()

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
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

			if line != "\r" {
				_, err = pipeWriter.Write([]byte(line))
				pipeWriter.Close()
				if err != nil {
					break
				}
			}

			if command.printLine && line != "" {
				fmt.Println(line)
			}
			outputBuffer.WriteString(line)
			break
		}

		if line != "\n" && !strings.Contains(line, "\r") {
			_, err = pipeWriter.Write([]byte(line))
			if err != nil {
				break
			}
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
	logger.Debugf("[exec] CMD: %s, OUTPUT: %s", c.String(), res)
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
