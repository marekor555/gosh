// gosh, shell made in go
// Copyright (C) 2024  MAREKOR555, contact: marekor555@interia.pl
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
// YOU MUST CREDIT ME IF YOU USE THE PROGRAM OR ANY PARTS OF IT
package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

//go:embed goshrc
var defaultGoshrc []byte

var (
	currentDir                          string
	err                                 error
	aliases                             = map[string]string{}
	aliasesInt                          = map[string]string{"clear": "clear -x", "ls": "ls ."}
	lastDir                             string
	cmdRunner, pipeRunner, cmdRunnerRes *exec.Cmd
	debug                               *bool
)

func parseTime(time int) string { // adds 0 if time is between 0 and 9
	if time < 10 {
		return "0" + strconv.Itoa(time)
	}
	return strconv.Itoa(time)
}

func main() {
	color.Yellow(`
	gosh  Copyright (C) 2024  MAREKOR555
	This program comes with ABSOLUTELY NO WARRANTY.
	This is free software, and you are welcome to redistribute it
	YOU MUST CREDIT ME IF YOU USE THE PROGRAM OR ANY PARTS OF IT

`)
	debug = flag.Bool("debug", false, "run debug mode")
	flag.Parse()
	if runtime.GOOS == "windows" {
		color.Red("no, just no")
		os.Exit(99)
	}
	command := ""
	reader := bufio.NewReader(os.Stdin) // initialize reader for getting user input
	currentDir, err = os.Getwd()        // get current working directory
	lastDir = currentDir
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

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, os.Interrupt)
	signal.Notify(osSig, syscall.SIGTERM)
	go func() {
		for {
			<-osSig
			if cmdRunner != nil {
				cmdRunner.Process.Kill()
			}
			if pipeRunner != nil {
				pipeRunner.Process.Kill()
			}
			if cmdRunnerRes != nil {
				cmdRunnerRes.Process.Kill()
			}
		}
	}()

	for {
		hi, mi, si := time.Now().Clock() // parse time
		h := parseTime(hi)
		m := parseTime(mi)
		s := parseTime(si)
		directories := strings.Split(currentDir, "/")
		dirName := directories[len(directories)-1]
		if dirName == "" {
			dirName = "/"
		}
		fmt.Print( // print prompt
			color.CyanString(fmt.Sprintf("%v:%v:%v ", h, m, s)),
			color.GreenString(user.Username),
			" at ",
			color.GreenString(dirName),
			color.BlueString(" >"), color.MagentaString(">"), color.BlueString("> "),
		)
		command, err = reader.ReadString('\n') // read user input
		if err != nil {
			color.Red("couldn't get user input:")
			color.Red(err.Error())
		}
		if strings.TrimSpace(command) == "" { // if command contains only spaces or is empty, go to prompt
			continue
		}
		commands := cmdSplit(command, "&&") // split input into commands
		for _, command := range commands {
			command = strings.TrimSpace(command)

			switch checkCustom(command) {
			case "cd":
				lastDir = currentDir
				newCurrentDir := currentDir
				if len(strings.Split(command, " ")) <= 1 { // if empty go to homedir
					currentDir = user.HomeDir
					break
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
					break
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
					break
				}
				dbg_print("New dir is:", newCurrentDir)
				currentDir = newCurrentDir
			case "uncd":
				dbg_print("Going back to: ", lastDir)
				currentDir = lastDir
			case "reloadCfg": // reload config command: reloadCfg
				loadConfig(user.HomeDir, reader)
				color.Green("Config reloaded")
			case "shellPath": // for debuging, shows path as seen by shell
				fmt.Println(currentDir)
			case "help":
				color.Blue("list of built in custom commands")
				color.Blue("help      - display help")
				color.Blue("cd        - unix cd wasn't compatible, so it is a custom command")
				color.Blue("uncd      - reverse last cd command")
				color.Blue("reloadCfg - reloads config from ~/.goshrc")
				color.Blue("shellPath - debug command to show shellPath variable (may be different than pwd)")
				color.Blue("exit      - exit shell")
			case "exit":
				os.Exit(0)
			default:
				if checkFor(command, "|") { // pipe command
					runPipe(command)
				} else if checkFor(command, ">>") { // redirect command
					runRedirect(command)
				} else if checkFor(command, "<<") { // redirect command
					runResult(command)
				} else {
					runCommand(command) // if no other command patterns match, it means that it is normal command
				}
			}
		}
	}
}
