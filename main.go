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
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

//go:embed .goshrc
var defaultGoshrc []byte

var (
	currentDir string
	err        error
	aliases    = map[string]string{}
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
	for key, value := range aliases {
		command = strings.Replace(command, key, value, 1)
	}
	cmdRunner := exec.Command("sh", "-c", command)
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
	if len(split) != 2 {
		color.Red("don't use | in piping commands elsewhere")
		return
	}
	cmd := split[0]
	pipe := split[1]
	for key, value := range aliases {
		cmd = strings.Replace(cmd, key, value, 1)
	}
	for key, value := range aliases {
		pipe = strings.ReplaceAll(pipe, key, value)
	}
	cmdRunner := exec.Command("sh", "-c", cmd)
	pipeRunner := exec.Command("sh", "-c", pipe)
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
	aliases["cls"] = "clear"
	aliases["ls"] = "lsd"
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
