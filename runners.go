package main

import (
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

func runRedirect(command string) {
	command = alias(command)            // alias command
	cmdSplit := cmdSplit(command, ">>") // split between command and file

	if len(cmdSplit) != 2 { // cannot redirect to more than two files
		color.Red("ERROR: more than one >> detected")
		return
	}

	cmd, args := parseCmd(cmdSplit[0])                                            // parse command
	file, err := os.Create(path.Join(currentDir, strings.TrimSpace(cmdSplit[1]))) // open file that will receive command stdout
	if err != nil {
		color.Red("Error: " + err.Error())
		return
	}

	cmdRunner = initCmd(cmd, args) // initialize command
	cmdRunner.Stdout = file        // redirect command stdout to file

	err = cmdRunner.Run() // start command and wait for finish
	if err != nil {
		color.Red("Error: " + err.Error())
	}
	cmdRunner = nil
}

func runResult(command string) {
	command = alias(command)
	cmdSplit := cmdSplit(command, "<<")

	if len(cmdSplit) != 2 {
		color.Red("ERROR: more than one << detected")
	}
	cmd, args := parseCmd(cmdSplit[0])
	cmdRes, argsRes := parseCmd(cmdSplit[1])

	cmdRunnerRes = initCmd(cmdRes, argsRes)

	output, err := cmdRunnerRes.Output()
	if err != nil {
		color.Red("Error: " + err.Error())
	}

	clean := strings.TrimSpace(string(output))
	args = append(args, strings.Split(clean, "\n")...)

	cmdRunner = initCmd(cmd, args)
	cmdRunner.Stdout = os.Stdout
	cmdRunner.Stderr = os.Stderr
	cmdRunner.Stdin = os.Stdin

	err = cmdRunner.Run()
	if err != nil {
		color.Red("Error: " + err.Error())
	}

	cmdRunner = nil
	cmdRunnerRes = nil
}

func runPipe(command string) {
	split := cmdSplit(command, "|") // split between pipe and command beeing piped
	if len(split) != 2 {
		color.Red("ERROR: more than one | detected")
		return
	}

	split[0] = alias(split[0]) // alias command
	split[1] = alias(split[1]) // alias pipe

	cmd, cmdArgs := parseCmd(split[0])   // parse command
	pipe, pipeArgs := parseCmd(split[1]) // parse pipe

	cmdRunner = initCmd(cmd, cmdArgs)    // init command
	pipeRunner = initCmd(pipe, pipeArgs) // init pipe

	cmdRunner.Stderr = os.Stderr  // command err is os.Stderr
	pipeRunner.Stderr = os.Stderr // pipe err is os.Stderr

	pipeRunner.Stdin, err = cmdRunner.StdoutPipe() // redirect stdout of command to stdin of pipe
	if err != nil {                                // check if piping failed
		color.Red("couldn't pipe command")
		color.Red(err.Error())
		return
	}

	pipeRunner.Stdout = os.Stdout // stdout of pipe is os.Stdout
	err = cmdRunner.Start()       // start command (and not wait)
	if err != nil {
		color.Red("failed to start command")
		color.Red(err.Error())
		return
	}

	err := pipeRunner.Start() // start pipe (and not wait)
	if err != nil {
		color.Red("failed to start piping command")
		color.Red(err.Error())
		return
	}

	err = cmdRunner.Wait() // wait for end of command
	if err != nil {
		color.Red("Main command error: " + err.Error())
	}

	err = pipeRunner.Wait() // wait for end of pipe
	if err != nil {
		color.Red("Pipe command error: " + err.Error())
	}

	cmdRunner = nil
	pipeRunner = nil
}

func runCommand(command string) {
	command = alias(command)       // alias command
	cmd, args := parseCmd(command) // parse command
	cmdRunner = initCmd(cmd, args) // init command

	cmdRunner.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmdRunner.Stdin = os.Stdin // make commands io be io of os
	cmdRunner.Stdout = os.Stdout
	cmdRunner.Stderr = os.Stderr

	err = cmdRunner.Run() // run command
	if err != nil {
		color.Red("Error: " + err.Error())
	}

	cmdRunner = nil
}
