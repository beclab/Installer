package util

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
)

func Exec(name string, printOutput bool, printLine bool) (stdout string, code int, err error) {
	exitCode := 0

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", name)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", exitCode, err
	}

	// logger.Infof("exec cmd: %s", cmd.String())
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return "", exitCode, err
	}

	var outputBuffer bytes.Buffer
	r := bufio.NewReader(out)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("[exec] read error: %s", err)
			}

			if printLine && line != "" {
				fmt.Println(strings.TrimSuffix(line, "\n"))
			}
			outputBuffer.WriteString(line)

			break
		}

		if printLine && line != "" {
			fmt.Println(strings.TrimSuffix(line, "\n"))
		}

		outputBuffer.WriteString(line)
	}

	err = cmd.Wait()
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	res := outputBuffer.String()
	res = strings.TrimSpace(res)

	if printOutput {
		fmt.Printf("[exec] CMD: %s, OUTPUT: \n%s\n", cmd.String(), res)
	}

	logger.Infof("[exec] CMD: %s, OUTPUT: %s", cmd.String(), res)
	return res, exitCode, errors.Wrapf(err, "Failed to exec command: %s \n%s", cmd, res)
}
