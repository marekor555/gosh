// gosh, shell made in go
// Copyright (C) 2024  MAREKOR555, contact: marekor555@interia.pl
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	currentDir string
	err        error
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
func runCommand(command string) {
	cmdSplit := strings.Split(command, " ")
	cmdRunner := exec.Command(cmdSplit[0], cmdSplit[1:]...)
	cmdRunner.Dir = currentDir
	cmdRunner.Stdin = os.Stdin
	cmdRunner.Stdout = os.Stdout
	cmdRunner.Stderr = os.Stderr
	err = cmdRunner.Run()
	if err != nil {
		color.Red("Error: " + err.Error())
	}
}
func runPipe(command string) {
	split := strings.Split(command, "|")
	if len(split) > 2 {
		color.Red("don't use | in piping commands elsewhere")
		return
	}
	cmdSplit := strings.Split(strings.TrimSpace(split[0]), " ")
	pipeSplit := strings.Split(strings.TrimSpace(split[1]), " ")
	cmdRunner := exec.Command(cmdSplit[0], cmdSplit[1:]...)
	pipeRunner := exec.Command(pipeSplit[0], pipeSplit[1:]...)
	pipeRunner.Stdin, err = cmdRunner.StdoutPipe()
	pipeRunner.Stdout = os.Stdout
	if err != nil {
		color.Red("couldn't pipe command")
		color.Red(err.Error())
	}
	pipeRunner.Start()
	cmdRunner.Run()
	pipeRunner.Wait()
}
func main() {
	color.Yellow(`
	gosh  Copyright (C) 2024  MAREKOR555
	This program comes with ABSOLUTELY NO WARRANTY.
	This is free software, and you are welcome to redistribute it
`)
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
	for {
	prompt:
		hi, mi, si := time.Now().Clock()
		h := parseTime(hi)
		m := parseTime(mi)
		s := parseTime(si)
		fmt.Print(color.CyanString(fmt.Sprintf("%v:%v:%v ", h, m, s)), color.GreenString(user.Username), color.BlueString(" >"), color.MagentaString(">"), color.BlueString("> "))
		command, err = reader.ReadString('\n')
		if err != nil {
			color.Red("couldn't get user input:")
			color.Red(err.Error())
		}
		commands := strings.Split(command, "&&")
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
			if strings.Contains(command, "|") {
				runPipe(command)
			} else {
				runCommand(command)
			}
		}
	}
}
