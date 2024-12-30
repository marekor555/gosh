// gosh, shell made in go
// Copyright (C) 2024  MAREKOR555, contact: marekor555@interia.pl
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
// YOU MUST CREDIT ME IF YOU USE THE PROGRAM OR ANY PARTS OF IT
package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

//go:embed goshrc
var defaultGoshrc []byte

var (
	currentDir string
	err        error
	aliases    = map[string]string{}
	aliasesInt = map[string]string{"clear": "clear -x"}
)

func checkCustom(text string, command string) bool { // check for custom command
	return strings.Contains(strings.Split(text, " ")[0], command)
}
func parseTime(time int) string { // adds 0 if time is between 0 and 9
	if time < 10 {
		return "0" + strconv.Itoa(time)
	}
	return strconv.Itoa(time)
}
func initCmd(command string, args []string) *exec.Cmd {
	var cmd *exec.Cmd
	if len(args) == 0 { // create command without args if len(args) is 0
		cmd = exec.Command(command)
	} else { // otherwise init with args
		cmd = exec.Command(command, args...)
	}
	cmd.Dir = currentDir   // current dir is command's dir
	cmd.Env = os.Environ() // enviromental variables are command's variables
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")
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
	split := []string{}
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
		if c == '"' || c == '\'' {
			quotes = !quotes
		}
		if !quotes {
			cmd += string(c)
		}
	}
	return strings.Contains(cmd, keyword)
}
func parseCmd(command string) (string, []string) { // just works
	cmd := ""                                     // command to be outputed
	args := []string{}                            // arguments
	commandSplit := strings.Fields(command)       // split with spaces
	cmd = commandSplit[0]                         // get command
	argsNP := strings.Join(commandSplit[1:], " ") // other parts will be arguments
	quotes := false
	arg := ""
	for _, c := range argsNP {
		if c == '"' || c == '\'' { // quote toggle, it means 'hello" is valid quote, to be fixed
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
func runCommand(command string) {
	command = alias(command)        // alias command
	cmd, args := parseCmd(command)  // parse command
	cmdRunner := initCmd(cmd, args) // init command
	cmdRunner.Stdin = os.Stdin      // make commands io be io of os
	cmdRunner.Stdout = os.Stdout
	cmdRunner.Stderr = os.Stderr
	err = cmdRunner.Run() // run command
	if err != nil {
		color.Red("Error: " + err.Error())
	}
}
func runRedirect(command string) {
	command = alias(command)            // alias command
	cmdSplit := cmdSplit(command, ">>") // split between command and file
	if len(cmdSplit) != 2 {             // cannot redirect to more than two files
		color.Red("ERROR: more than one >> detected")
		return
	}
	cmd, args := parseCmd(cmdSplit[0])                     // parse command
	file, err := os.Create(strings.TrimSpace(cmdSplit[1])) // open file that will receive command stdout
	if err != nil {
		color.Red("Error: " + err.Error())
		return
	}
	cmdRunner := initCmd(cmd, args) // initialize command
	cmdRunner.Stdout = file         // redirect command stdout to file
	err = cmdRunner.Run()           // start command and wait for finish
	if err != nil {
		color.Red("Error: " + err.Error())
	}
}
func runPipe(command string) {
	split := cmdSplit(command, "|") // split between pipe and command beeing piped
	if len(split) != 2 {
		color.Red("ERROR: more than one | detected")
		return
	}
	split[0] = alias(split[0])                     // alias command
	split[1] = alias(split[1])                     // alias pipe
	cmd, cmdArgs := parseCmd(split[0])             // parse command
	pipe, pipeArgs := parseCmd(split[1])           // parse pipe
	cmdRunner := initCmd(cmd, cmdArgs)             // init command
	pipeRunner := initCmd(pipe, pipeArgs)          // init pipe
	cmdRunner.Stderr = os.Stderr                   // command err is os.Stderr
	pipeRunner.Stderr = os.Stderr                  // pipe err is os.Stderr
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
}
func loadConfig(homedir string, reader *bufio.Reader) {
	file, err := os.Open(path.Join(homedir, ".goshrc")) // open config file
	if err != nil {
		color.Red("failed to open ~/.goshrc")
		color.Red(err.Error())
		color.Red("if config isn't initialzed, click enter to create default ~/.goshrc")
		reader.ReadString('\n')
		file.Close()
		file, err := os.Create(path.Join(homedir, ".goshrc"))
		if err != nil {
			color.Red("failed to create ~/.goshrc")
			color.Red(err.Error())
		}
		defer file.Close()
		_, err = file.Write(defaultGoshrc)
		if err != nil {
			color.Red("failed to write to ~/.goshrc")
			color.Red(err.Error())
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file) // scan file
	for scanner.Scan() {
		if scanner.Text()[0] == '#' { // comments start with #
			continue
		}
		splitText := strings.Split(scanner.Text(), ">>>") // alias keyword: >>>
		aliases[splitText[0]] = splitText[1]
	}
}
func main() {
	color.Yellow(`
	gosh  Copyright (C) 2024  MAREKOR555
	This program comes with ABSOLUTELY NO WARRANTY.
	This is free software, and you are welcome to redistribute it
	YOU MUST CREDIT ME IF YOU USE THE PROGRAM OR ANY PARTS OF IT

`)
	if runtime.GOOS == "windows" {
		color.Red("no, just no")
		os.Exit(99)
	}
	command := ""
	reader := bufio.NewReader(os.Stdin) // initialize reader for getting user input
	currentDir, err = os.Getwd()        // get current working directory
	if err != nil {
		color.Red("couldn't get working directory")
		color.Red(err.Error())
	}
	user, err := user.Current() // get current user
	if err != nil {
		color.Red("couldn't get current user")
		color.Red(err.Error())
	}
	loadConfig(user.HomeDir, reader) // load config
	for {
	prompt:
		hi, mi, si := time.Now().Clock() // parse time
		h := parseTime(hi)
		m := parseTime(mi)
		s := parseTime(si)
		fmt.Print( // print prompt
			color.CyanString(fmt.Sprintf("%v:%v:%v ", h, m, s)),
			color.GreenString(user.Username),
			color.BlueString(" >"), color.MagentaString(">"), color.BlueString("> "),
		)
		command, err = reader.ReadString('\n') // read user input
		if err != nil {
			color.Red("couldn't get user input:")
			color.Red(err.Error())
		}
		if strings.TrimSpace(command) == "" { // if command contains only spaces or is empty, go to prompt
			goto prompt
		}
		commands := cmdSplit(command, "&&") // split input into commands
		for _, command := range commands {
			command = strings.TrimSpace(command)
			if checkCustom(command, "cd") {
				newCurrentDir := currentDir
				if len(strings.Split(command, " ")) <= 1 { // if empty go to homedir
					currentDir = user.HomeDir
					goto prompt
				}
				currentDir = strings.ReplaceAll(currentDir, "~", user.HomeDir) // replace ~ with user.Homedir
				if strings.Count(command, "..") > 0 {                          // check if there is .. and remove directories from path
					backCount := strings.Count(command, "..")
					currentDirSplit := strings.Split(newCurrentDir, "/")
					if backCount >= len(currentDirSplit) {
						backCount = len(currentDirSplit) - 1
					}
					currentDirSplit = currentDirSplit[:len(currentDirSplit)-backCount] // split path and ignore directories that should be deleted from path
					newCurrentDir = ""                                                 // clear new current directory path
					for _, dir := range currentDirSplit {                              // join the directories
						newCurrentDir += "/" + dir
					}
				}
				if strings.Split(command, " ")[1][0] == '/' { // if first char of path is / then use this path, it's absolute
					currentDir = strings.Split(command, " ")[1]
					goto prompt
				}
				for _, dir := range strings.Split(strings.Split(command, " ")[1], "/") { // add non relative path, and ignore .., it was replaced before
					if dir == ".." {
						continue
					}
					newCurrentDir += "/" + dir
				}
				newCurrentDir = strings.ReplaceAll(newCurrentDir, "//", "/") // replace // with / when if there is an error with that
				if _, err := os.Stat(newCurrentDir); os.IsNotExist(err) {
					color.Red("Error: path doesn't exist")
					goto prompt
				}
				currentDir = newCurrentDir
				goto prompt
			}
			if checkCustom(command, "reloadCfg") { // reload config command: reloadCfg
				loadConfig(user.HomeDir, reader)
				color.Green("Config reloaded")
				goto prompt
			}
			if checkCustom(command, "shellPath") { // for debuging, shows path as seen by shell
				fmt.Println(currentDir)
				goto prompt
			}
			if checkCustom(command, "help") {
				color.Blue("list of built in custom commands")
				color.Blue("help      - display help")
				color.Blue("cd        - unix cd wasn't compatible, so it is a custom command")
				color.Blue("reloadCfg - reloads config from ~/.goshrc")
				color.Blue("shellPath - debug command to show shellPath variable (may be different than pwd)")
				color.Blue("exit      - exit shell")
				goto prompt
			}
			if checkCustom(command, "exit") {
				os.Exit(0)
			}
			if checkFor(command, "|") { // pipe command
				runPipe(command)
			} else if checkFor(command, ">>") { // redirect command
				runRedirect(command)
			} else {
				runCommand(command) // if no other command patterns match, it means that it is normal command
			}
		}
	}
}
