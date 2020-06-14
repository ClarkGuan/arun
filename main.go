package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
		os.Exit(1)
	}

	targetFile := os.Args[1]
	isTest := false
	otherArgs := os.Args[2:]
	switch os.Args[1] {
	case "-exe", "exe":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
			os.Exit(1)
		}
		targetFile = os.Args[2]
		otherArgs = os.Args[3:]

	case "-test", "test":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
			os.Exit(1)
		}
		targetFile = os.Args[2]
		isTest = true
		otherArgs = os.Args[3:]
	}

	execFile, err := filepath.Abs(targetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not found\n", execFile)
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("prepare to push %s to device\n", execFile)

	if err := runCmd("adb", "push", execFile, "/data/local/tmp/"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var args []string
	args = append(args, "shell",
		"time",
		"sh", "-c",
		"'",
		"echo \"[程序输出如下]\" && LD_LIBRARY_PATH=/data/local/tmp",
		"/data/local/tmp/"+filepath.Base(execFile))
	if isTest {
		found := false
		for _, arg := range otherArgs {
			if strings.HasPrefix(arg, "-test.v") {
				found = true
				break
			}
		}
		if !found {
			args = append(args, "-test.v=true")
		}
	}
	args = append(args, otherArgs...)
	args = append(args, "&& echo \"[程序执行完毕]\" || echo \"[程序执行返回错误码($?)]\"")
	args = append(args, "&& rm "+"/data/local/tmp/"+filepath.Base(execFile))
	args = append(args, "'")
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
