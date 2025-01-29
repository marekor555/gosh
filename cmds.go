package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
)

func checkCustom(text string) string { // check for custom command
	return strings.Split(text, " ")[0]
}

func loadConfig(homedir string, reader *bufio.Reader) {
	file, err := os.Open(path.Join(homedir, ".goshrc")) // open config file

	if err != nil {
		color.Red("failed to open ~/.goshrc")
		color.Red(err.Error())
		color.Red("if config isn't initialized, click enter to create default ~/.goshrc")

		_, _ = reader.ReadString('\n')
		err := file.Close()
		if err != nil {
			return
		}

		file, err := os.Create(path.Join(homedir, ".goshrc"))
		if err != nil {
			color.Red("failed to create ~/.goshrc")
			color.Red(err.Error())
		}
		defer func() {
			_ = file.Close()
		}()

		_, err = file.Write(defaultGoshrc)
		if err != nil {
			color.Red("failed to write to ~/.goshrc")
			color.Red(err.Error())
		}
	}
	defer func() {
		_ = file.Close()
	}()
	data, _ := os.ReadFile(path.Join(homedir, ".goshrc"))
	dbgPrint("Loading: \n", string(data))
	scanner := bufio.NewScanner(file) // scan file

	for scanner.Scan() {
		if scanner.Text()[0] == '#' { // comments start with #
			continue
		}
		splitText := strings.Split(scanner.Text(), ">>>") // alias keyword: >>>
		aliases[splitText[0]] = splitText[1]
	}
}

func initCmd(command string, args []string) *exec.Cmd {
	var cmd *exec.Cmd

	if len(args) == 0 || (len(args) == 1 && args[0] == "") { // create command without args if len(args) is 0 or the only arg is empty
		cmd = exec.Command(command)
	} else { // otherwise init with args
		cmd = exec.Command(command, args...)
	}

	cmd.Dir = currentDir   // current dir is command's dir
	cmd.Env = os.Environ() // environmental variables are command's variables
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	dbgPrint("Cmd:         ", command)
	dbgPrint("Args amount: ", len(args))
	dbgPrint("Args:        ", args)

	return cmd
}

func alias(command string) string { // replace aliases + internal quickfixes
	commandSplit := strings.Fields(command)

	for key, val := range aliases {
		if commandSplit[0] == key {
			command = strings.Replace(command, key, val, 1)
		}
	}

	for key, val := range aliasesInt {
		if commandSplit[0] == key {
			command = strings.Replace(command, key, val, 1)
		}
	}

	return command
}

func cmdSplit(command string, splitKey string) []string {
	var split []string
	element := ""

	for _, c := range command {
		element += string(c)

		if checkFor(element, splitKey) {
			elementSplit := strings.Split(element, splitKey)
			element = strings.Join(elementSplit[:len(elementSplit)-1], splitKey)
			split = append(split, element)
			element = ""
		}
	}

	split = append(split, element)
	return split
}

func checkFor(command string, keyword string) bool {
	quotes := false
	cmd := ""
	for _, c := range command { // ignoring everything in quotes, it is not command
		if c == '"' {
			quotes = !quotes
		}

		if !quotes {
			cmd += string(c)
		}
	}

	return strings.Contains(cmd, keyword)
}

func parseCmd(command string) (string, []string) { // just works
	cmd := ""                                     // command to be outputted
	var args []string                             // arguments
	commandSplit := strings.Fields(command)       // split with spaces
	cmd = commandSplit[0]                         // get command
	argsNP := strings.Join(commandSplit[1:], " ") // other parts will be arguments

	quotes := false
	arg := ""

	for _, c := range argsNP {
		if c == '"' { // quote toggle, it means 'hello" is valid quote, to be fixed
			quotes = !quotes
			continue
		}

		if !quotes && c == ' ' { // if not in quotes and space, new argument is created
			args = append(args, arg)
			arg = ""
			continue
		}

		arg += string(c) // add character to argument
	}

	args = append(args, arg) // append last argument
	return cmd, args
}

func dbgPrint(msg string, vars ...any) {
	if *debug {
		fmt.Print(color.RedString(msg))
		fmt.Println(vars...)
	}
}
