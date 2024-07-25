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

func checkCustom(text string, command string) bool {
	return strings.Contains(strings.Split(text, " ")[0], command)
}
func parseTime(time int) string {
	if time < 10 {
		return "0" + strconv.Itoa(time)
	}
	return strconv.Itoa(time)
}
func initCmd(command string, args []string) *exec.Cmd {
	var cmd *exec.Cmd
	if len(args) == 0 {
		cmd = exec.Command(command)
	} else {
		cmd = exec.Command(command, args...)
	}
	cmd.Env = os.Environ()
	return cmd
}
func alias(command string) string {
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
	for _, c := range command {
		if c == '"' || c == '\'' {
			quotes = !quotes
		}
		if !quotes {
			cmd += string(c)
		}
	}
	return strings.Contains(cmd, keyword)
}
func parseCmd(command string) (string, []string) {
	cmd := ""
	args := []string{}
	commandSplit := strings.Fields(command)
	cmd = commandSplit[0]
	argsNP := strings.Join(commandSplit[1:], " ")
	quotes := false
	arg := ""
	for _, c := range argsNP {
		if c == '"' || c == '\'' {
			quotes = !quotes
			continue
		}
		if !quotes && c == ' ' {
			args = append(args, arg)
			arg = ""
			continue
		}
		arg += string(c)
	}
	args = append(args, arg)
	return cmd, args
}
func runCommand(command string) {
	command = alias(command)
	cmd, args := parseCmd(command)
	cmdRunner := initCmd(cmd, args)
	cmdRunner.Dir = currentDir
	cmdRunner.Stdin = os.Stdin
	cmdRunner.Stdout = os.Stdout
	cmdRunner.Stderr = os.Stderr
	err = cmdRunner.Run()
	if err != nil {
		color.Red("Error: " + err.Error())
	}
}
func runRedirect(command string) {
	command = alias(command)
	cmdSplit := cmdSplit(command, ">>")
	if len(cmdSplit) != 2 {
		color.Red("ERROR: more than one >> detected")
		return
	}
	cmd, args := parseCmd(cmdSplit[0])
	file, err := os.Create(strings.TrimSpace(cmdSplit[1]))
	if err != nil {
		color.Red("Error: " + err.Error())
		return
	}
	cmdRunner := initCmd(cmd, args)
	cmdRunner.Stdout = file
	err = cmdRunner.Run()
	if err != nil {
		color.Red("Error: " + err.Error())
	}
}
func runPipe(command string) {
	split := cmdSplit(command, "|")
	if len(split) != 2 {
		color.Red("ERROR: more than one | detected")
		return
	}
	split[0] = alias(split[0])
	split[1] = alias(split[1])
	cmd, cmdArgs := parseCmd(split[0])
	pipe, pipeArgs := parseCmd(split[1])
	cmdRunner := initCmd(cmd, cmdArgs)
	pipeRunner := initCmd(pipe, pipeArgs)
	cmdRunner.Stderr = os.Stderr
	pipeRunner.Stderr = os.Stderr
	pipeRunner.Stdin, err = cmdRunner.StdoutPipe()
	if err != nil {
		color.Red("couldn't pipe command")
		color.Red(err.Error())
		return
	}
	pipeRunner.Stdout = os.Stdout
	err = cmdRunner.Start()
	if err != nil {
		color.Red("failed to start command")
		color.Red(err.Error())
		return
	}
	err := pipeRunner.Start()
	if err != nil {
		color.Red("failed to start piping command")
		color.Red(err.Error())
		return
	}
	err = cmdRunner.Wait()
	if err != nil {
		color.Red("Main command error: " + err.Error())
	}
	err = pipeRunner.Wait()
	if err != nil {
		color.Red("Pipe command error: " + err.Error())
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
		color.Red("using gosh on windows isn't recommended, consider using powershell")
	}
	command := ""
	reader := bufio.NewReader(os.Stdin)
	currentDir, err = os.Getwd()
	if err != nil {
		color.Red("couldn't get working directory")
		color.Red(err.Error())
	}
	user, err := user.Current()
	if err != nil {
		color.Red("couldn't get current user")
		color.Red(err.Error())
	}
	file, err := os.Open(path.Join(user.HomeDir, ".goshrc"))
	if err != nil {
		color.Red("failed to open ~/.goshrc")
		color.Red(err.Error())
		color.Red("if config isn't initialzed, click enter to create default ~/.goshrc")
		reader.ReadString('\n')
		file.Close()
		file, err := os.Create(path.Join(user.HomeDir, ".goshrc"))
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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text()[0] == '#' {
			continue
		}
		splitText := strings.Split(scanner.Text(), ">>>")
		aliases[splitText[0]] = splitText[1]
	}
	for {
	prompt:
		hi, mi, si := time.Now().Clock()
		h := parseTime(hi)
		m := parseTime(mi)
		s := parseTime(si)
		fmt.Print(
			color.CyanString(fmt.Sprintf("%v:%v:%v ", h, m, s)),
			color.GreenString(user.Username),
			color.BlueString(" >"), color.MagentaString(">"), color.BlueString("> "),
		)
		command, err = reader.ReadString('\n')
		if err != nil {
			color.Red("couldn't get user input:")
			color.Red(err.Error())
		}
		if strings.TrimSpace(command) == "" {
			goto prompt
		}
		commands := cmdSplit(command, "&&")
		for _, command := range commands {
			command = strings.TrimSpace(command)
			if checkCustom(command, "cd") {
				if len(strings.Split(command, " ")) <= 1 {
					currentDir = user.HomeDir
					goto prompt
				}
				if strings.Count(command, "..") > 0 {
					backCount := strings.Count(command, "..")
					currentDirSplit := strings.Split(currentDir, "/")
					if backCount >= len(currentDirSplit) {
						backCount = len(currentDirSplit) - 1
					}
					currentDirSplit = currentDirSplit[:len(currentDirSplit)-backCount]
					currentDir = ""
					for _, dir := range currentDirSplit {
						currentDir += "/" + dir
					}
					goto prompt
				}
				if strings.Split(command, " ")[1] == "~" {
					currentDir = user.HomeDir
					goto prompt
				}
				currentDir = strings.Split(command, " ")[1]
				goto prompt
			}
			if checkCustom(command, "shellPath") {
				fmt.Println(currentDir)
				goto prompt
			}
			if checkCustom(command, "exit") {
				os.Exit(0)
			}
			if runtime.GOOS == "windows" {
				command = alias(command)
				cmd := exec.Command("cmd", "/c", command)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					color.Red("Error: " + err.Error())
				}
				goto prompt
			}
			if checkFor(command, "|") {
				runPipe(command)
			} else if checkFor(command, ">>") {
				runRedirect(command)
			} else {
				runCommand(command)
			}
		}
	}
}
