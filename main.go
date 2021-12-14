package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	errNotELF = errors.New("target is not an ELF file")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
		os.Exit(1)
	}

	targetFile := os.Args[1]
	otherArgs := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "-exe", "exe":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
			os.Exit(1)
		}
		targetFile = os.Args[2]
		otherArgs = os.Args[3:]
		err = runExec(targetFile, otherArgs)

	case "-gotest", "gotest":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "缺少 可执行文件路径 这一必须参数")
			os.Exit(1)
		}
		targetFile = os.Args[2]
		otherArgs = os.Args[3:]
		err = runGoTest(targetFile, otherArgs)

	case "-jar", "jar":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "缺少 jar包 & 主类名 等必须参数")
			os.Exit(1)
		}
		targetFile = os.Args[2]
		otherArgs = os.Args[3:]
		err = runDexJar(targetFile, otherArgs)

	default:
		if isZip(targetFile) {
			if len(otherArgs) == 0 {
				fmt.Fprintln(os.Stderr, "缺少 主类名 这一必须参数")
				os.Exit(1)
			}
			err = runDexJar(targetFile, otherArgs)
		} else {
			err = runExec(targetFile, otherArgs)
		}
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runExec(target string, oArgs []string) error {
	if !isELF(target) {
		return errNotELF
	}

	execFile, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	fmt.Printf("prepare to push %s to device\n", execFile)
	if err := runCmd("adb", "push", execFile, "/data/local/tmp/"); err != nil {
		return err
	}
	var args []string
	fileTargetPath := "/data/local/tmp/" + filepath.Base(execFile)
	args = append(args, "shell",
		"cd /data/local/tmp/",
		"&& chmod +x", fileTargetPath,
		"&& echo \"[程序输出如下]\" &&",
		"time",
		"sh", "-c",
		"'",
		"LD_LIBRARY_PATH=/data/local/tmp",
		fileTargetPath)
	args = append(args, oArgs...)
	args = append(args, "&& echo \"[程序执行完毕]\" || echo \"[程序执行返回错误码($?)]\"",
		"'",
		"&& rm "+fmt.Sprintf("\"%s\"", fileTargetPath))
	err = runCmd("adb", args...)
	return err
}

func runGoTest(target string, oArgs []string) error {
	if !isELF(target) {
		return errNotELF
	}

	execFile, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	fmt.Printf("prepare to push %s to device\n", execFile)
	if err := runCmd("adb", "push", execFile, "/data/local/tmp/"); err != nil {
		return err
	}
	var args []string
	fileTargetPath := "/data/local/tmp/" + filepath.Base(execFile)
	args = append(args, "shell",
		"cd /data/local/tmp/",
		"&& chmod +x", fileTargetPath,
		"&& echo \"[程序输出如下]\" &&",
		"time",
		"sh", "-c",
		"'",
		"LD_LIBRARY_PATH=/data/local/tmp",
		fileTargetPath)

	found := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-test.v") {
			found = true
			break
		}
	}
	if !found {
		args = append(args, "-test.v=true")
	}

	args = append(args, oArgs...)
	args = append(args, "&& echo \"[程序执行完毕]\" || echo \"[程序执行返回错误码($?)]\"",
		"'",
		"&& rm "+fmt.Sprintf("\"%s\"", fileTargetPath))
	err = runCmd("adb", args...)
	return err
}

func runDexJar(target string, oArgs []string) error {
	execFile, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	fmt.Printf("prepare to push %s to device\n", execFile)
	if err := runCmd("adb", "push", execFile, "/data/local/tmp/"); err != nil {
		return err
	}
	var args []string
	args = append(args, "shell",
		"cd /data/local/tmp/",
		"&& echo \"[程序输出如下]\" &&",
		"time",
		"sh", "-c",
		"'",
		"LD_LIBRARY_PATH=/data/local/tmp",
		fmt.Sprintf("CLASSPATH=\"/data/local/tmp/%s\"", filepath.Base(execFile)),
		"app_process",
		"/")
	args = append(args, oArgs...)
	args = append(args, "&& echo \"[程序执行完毕]\" || echo \"[程序执行返回错误码($?)]\"",
		"'",
		"&& rm "+fmt.Sprintf("\"/data/local/tmp/%s\"", filepath.Base(execFile)))
	err = runCmd("adb", args...)
	return err
}

func runCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	return command.Run()
}

func isZip(file string) bool {
	buf := make([]byte, 2)
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return false
	}
	return buf[0] == 0x50 && buf[1] == 0x4B
}

func isELF(file string) bool {
	buf := make([]byte, 4)
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return false
	}
	return buf[0] == 0x7F && buf[1] == 0x45 && buf[2] == 0x4C && buf[3] == 0x46
}
