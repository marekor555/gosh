# ABOUT
Gosh is a simple shell for unix systems.
It supports windows in so called 'fallback mode' using cmd.exe in background.
# INSTALLATION
We will be compiling the project from source.

1. You firstly must clone the git repo: `git clone https://github.com/marekor555/gosh`
2. Then move to the project: `cd gosh`
3. Get the libraries: `go get .`
4. Install the project: `go install .`
3. Now try the app: `gosh --save --as "duckduckgo.html" "duckduckgo.com"`

If the above command fails, saying that gosh command is not found. 
You need to add go bin to path. And try again.
