## Prerequisites
To contribute code changes to this project you will need the following development kits.
 * [Go](https://golang.org/doc/install)
 * [Docker](https://docs.docker.com/engine/installation/)
 
Dockwatch uses Go modules. Use the Go version defined in [go.mod](go.mod) (currently Go 1.25).
You can check your current version of the go language as follows:
```bash
  ~ $ go version
  go version go1.25.0 linux/amd64
```


## Checking out the code
Do not place your code in the go source path.
```bash
git clone git@github.com:<yourfork>/dockwatch.git
cd dockwatch
```

## Building and testing
dockwatch is a go application and is built with go commands. The following commands assume that you are at the root level of your repo.
```bash
go build                               # compiles and packages an executable binary, dockwatch
go test ./... -v                       # runs tests with verbose output
./dockwatch                           # runs the application (outside of a container)
```

To build a Dockwatch image of your own, use the self-contained Dockerfiles. As the main Dockerfile, they can be found in `dockerfiles/`:
- `dockerfiles/Dockerfile.dev-self-contained` will build an image based on your current local Dockwatch files.
- `dockerfiles/Dockerfile.self-contained` will build an image based on current Dockwatch's repository on GitHub.

e.g.:
```bash
sudo docker build . -f dockerfiles/Dockerfile.dev-self-contained -t fugginold/dockwatch # to build an image from local files
```