package adventure

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
)

const UserInteractionDelimiter = "> "

var delimiters = [][]byte{
	[]byte(UserInteractionDelimiter),
	[]byte("File name: "),
}

const (
	saveDialogFirstPart  = "I can suspend your Adventure for you so that you can resume later"
	saveDialogSecondPart = "File name:"
)

type Adventure struct {
	cmd *exec.Cmd

	cmdIn  io.WriteCloser
	cmdOut io.Reader
	cmdErr io.Reader

	quitOut    chan struct{}
	output     chan string
	outputUser chan string
	readError  chan error

	quitErr   chan struct{}
	errOutput chan string

	saveDialogTicks chan struct{}
}

func new(cmd *exec.Cmd) (adv *Adventure, err error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	adv = &Adventure{
		cmd:    cmd,
		cmdIn:  stdin,
		cmdOut: stdout,
		cmdErr: stderr,
	}

	return
}

func New(executable string) (*Adventure, error) {
	return new(exec.Command(executable))
}

func Resume(executable string, saveFile string) (*Adventure, error) {
	return new(exec.Command(executable, "-r", saveFile))
}

func NewOrResume(executable string, saveFile string) (*Adventure, error) {
	if info, err := os.Stat(saveFile); err != nil && !info.IsDir() {
		return Resume(executable, saveFile)
	}
	return New(executable)
}

func startsWithDelimiter(data []byte, delimiter []byte) bool {
	if len(data) < len(delimiter) {
		return false
	}
	for i := range delimiter {
		if data[i] != delimiter[i] {
			return false
		}
	}
	return true
}

func splitOutput(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := range data {
		for _, delim := range delimiters {
			if startsWithDelimiter(data[i:], delim) {
				return i + len(delim) + 1, data[:i+len(delim)], nil
			}
		}
	}

	if atEOF {
		return len(data), data, io.EOF
	}

	return 0, nil, nil
}

func (adv *Adventure) outputLoop() {
	defer close(adv.output)
	defer close(adv.readError)

	scanner := bufio.NewScanner(adv.cmdOut)
	scanner.Split(splitOutput)

	for scanner.Scan() {
		output := scanner.Text()

		select {
		case adv.output <- output:
		case <-adv.quitOut:
			break
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		adv.readError <- err
	}
}

func (adv *Adventure) errLoop() {
	defer close(adv.errOutput)

	scanner := bufio.NewScanner(adv.cmdErr)

	for scanner.Scan() {
		output := scanner.Text()
		output = strings.Trim(output, " \t\r\n")
		if output == "" {
			continue
		}

		select {
		case adv.errOutput <- output:
		case <-adv.quitErr:
			break
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		adv.readError <- err
	}
}

func (adv *Adventure) copyOutputAndLookForSaveDialog() {
	defer close(adv.outputUser)
	defer close(adv.saveDialogTicks)

	for output := range adv.output {
		if strings.Contains(output, saveDialogFirstPart) || strings.Contains(output, saveDialogSecondPart) {
			adv.saveDialogTicks <- struct{}{}
		} else {
			adv.outputUser <- output
		}
	}
}

func (adv *Adventure) Start() (output <-chan string, errorOutput <-chan string, err error) {
	adv.quitOut = make(chan struct{})
	adv.quitErr = make(chan struct{})

	adv.errOutput = make(chan string)
	errorOutput = adv.errOutput
	adv.output = make(chan string)
	adv.outputUser = make(chan string)
	output = adv.outputUser
	adv.readError = make(chan error, 1)

	go adv.copyOutputAndLookForSaveDialog()
	go adv.outputLoop()
	go adv.errLoop()

	err = adv.cmd.Start()

	if err != nil {
		adv.cleanUp()
	}

	return
}

func (adv *Adventure) Output() (output <-chan string) {
	return adv.outputUser
}

func (adv *Adventure) Error() (errorOutput <-chan string) {
	return adv.errOutput
}

func (adv *Adventure) Writeln(text string) error {
	_, err := adv.cmdIn.Write([]byte(text + "\n"))
	return err
}

func (adv *Adventure) SaveAndClose(saveFile string) error {
	adv.Writeln("save")
	<-adv.saveDialogTicks
	adv.Writeln("yes")
	<-adv.saveDialogTicks
	adv.Writeln(saveFile)

	// TODO: Check stderr for errors

	return adv.Close()
}

func (adv *Adventure) cleanUp() {
	close(adv.quitErr)
	close(adv.quitOut)
}

func (adv *Adventure) Close() error {
	adv.cleanUp()

	// Stop the command by closing the stdin
	adv.cmdIn.Close()
	return adv.cmd.Wait()
}
