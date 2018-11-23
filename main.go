package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func main() {
	var mode string
	var path string

	flag.StringVar(&mode, "m", "debug", "debug, release or relwithdebinfo...")
	flag.StringVar(&path, "p", ".", "project path")
	flag.Parse()

	cmakeBuildDir := filepath.Join(path, "cmake-build-"+mode)
	if _, err := os.Stat(cmakeBuildDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s not found\n", cmakeBuildDir)
		os.Exit(1)
	}

	cmakeBuildFile := filepath.Join(path, "CMakeLists.txt")
	targets := make([][]byte, 0, 8)
	if content, err := ioutil.ReadFile(cmakeBuildFile); err != nil {
		fmt.Fprintf(os.Stderr, "%s open fail", cmakeBuildFile)
		os.Exit(1)
	} else {
		regexExpr := regexp.MustCompile(`add_executable\s*\(\s*(\w+)\s+`)
		groups := regexExpr.FindSubmatch(content)
		if len(groups) > 1 {
			targets = append(targets, groups[1:]...)
		}
	}
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "cannot find any add_executable target")
		os.Exit(1)
	}
	//fmt.Println("find target(s) list below:")
	//for i, target := range targets {
	//	fmt.Printf("%d: %s\n", i, target)
	//}
	//fmt.Printf("please enter index: ")
	//
	//var index int
	//fmt.Scanf("%d", &index)
	//
	//if index >= len(targets) {
	//	fmt.Fprintln(os.Stderr, "index enterd out of bands")
	//	os.Exit(1)
	//}

	index := 0

	execFile, err := filepath.Abs(filepath.Join(cmakeBuildDir, string(targets[index])))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not found\n", execFile)
		os.Exit(1)
	}
	fmt.Printf("prepare to push %s to device\n", execFile)

	if err := runCmd("adb", "push", execFile, "/data/local/tmp/"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var args []string
	args = append(args, "shell",
		"echo \"[程序输出如下]\" && LD_LIBRARY_PATH=/data/local/tmp",
		"/data/local/tmp/"+filepath.Base(execFile))
	args = append(args, flag.Args()...)
	args = append(args, "&& echo \"[程序执行完毕]\"")
	if err := runCmd("adb", args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	return command.Run()
}
