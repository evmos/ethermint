// +build none

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	GOLANGCI_VERSION = "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.23.8"
)

func main() {

	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("scripts", "ci.go")); os.IsNotExist(err) {
		log.Fatal("should run build from root dir")
	} else if err != nil {
		panic(err)
	}
	if len(os.Args) < 2 {
		log.Fatal("cmd required, eg: install")
	}
	switch os.Args[1] {
	case "lint":
		lint()
	default:
		log.Fatal("cmd not found", os.Args[1])
	}
}

func lint() {

	verbose := flag.Bool("v", false, "Whether to log verbosely")

	// Make sure golangci-lint is available
	argsGet := append([]string{"get", GOLANGCI_VERSION})
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), argsGet...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("could not list packages: %v\n%s", err, string(out))
	}

	cmd = exec.Command(filepath.Join(GOBIN(), "golangci-lint"))
	cmd.Args = append(cmd.Args, "run", "--config", ".golangci.yml")

	if *verbose {
		cmd.Args = append(cmd.Args, "-v")
	}

	fmt.Println("Lint Ethermint", strings.Join(cmd.Args, " \\\n"))
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatal("Error: Could not Lint Ethermint. ", "error: ", err, ", cmd: ", cmd)
	}
}

// GOBIN returns the GOBIN environment variable
func GOBIN() string {
	if os.Getenv("GOBIN") == "" {
		log.Fatal("GOBIN is not set")
	}
	return os.Getenv("GOBIN")
}
