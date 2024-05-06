package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com//ShweetShell/Project2/builtins" // Updated import path
)

func main() {
	exit := make(chan struct{}, 2) // buffer this so there's no deadlock.
	runLoop(os.Stdin, os.Stdout, os.Stderr, exit)
}

func runLoop(r io.Reader, w, errW io.Writer, exit chan struct{}) {
	var (
		input    string
		err      error
		readLoop = bufio.NewReader(r)
	)
	for {
		select {
		case <-exit:
			_, _ = fmt.Fprintln(w, "exiting gracefully...")
			return
		default:
			if err := printPrompt(w); err != nil {
				_, _ = fmt.Fprintln(errW, err)
				continue
			}
			if input, err = readLoop.ReadString('\n'); err != nil {
				_, _ = fmt.Fprintln(errW, err)
				continue
			}
			if err = handleInput(w, input, exit); err != nil {
				_, _ = fmt.Fprintln(errW, err)
			}
		}
	}
}

func printPrompt(w io.Writer) error {
	// Get current user.
	// Don't prematurely memoize this because it might change due to `su`?
	u, err := user.Current()
	if err != nil {
		return err
	}
	// Get current working directory.
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// /home/User [Username] $
	_, err = fmt.Fprintf(w, "%v [%v] $ ", wd, u.Username)

	return err
}

func handleInput(w io.Writer, input string, exit chan<- struct{}) error {
	//  trailing spaces.
	input = strings.TrimSpace(input)

	args := strings.Split(input, " ")
	name, args := args[0], args[1:]

	//built-in commands.
	// New builtin commands should be added here. Eventually this should be refactored to its own func.
	switch name {
	case "cd":
		return builtins.ChangeDirectory(args...)
	case "env":
		return builtins.EnvironmentVariables(w, args...)
	case "exit":
		exit <- struct{}{}
		return nil
	case "echo":
		_, err := fmt.Fprintln(w, strings.Join(args, " "))
		return err
	case "pwd":
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, wd)
		return err
	case "mkdir":
		if len(args) < 1 {
			return fmt.Errorf("mkdir: missing operand")
		}
		for _, dir := range args {
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}
		return nil
	case "rmdir":
		if len(args) < 1 {
			return fmt.Errorf("rmdir: missing operand")
		}
		for _, dir := range args {
			err := os.Remove(dir)
			if err != nil {
				return err
			}
		}
		return nil
	case "touch":
		if len(args) < 1 {
			return fmt.Errorf("touch: missing operand")
		}
		for _, file := range args {
			f, err := os.Create(file)
			if err != nil {
				return err
			}
			f.Close()
		}
		return nil
	}

	return executeCommand(name, args...)
}

func executeCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	// the output device setting
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}
