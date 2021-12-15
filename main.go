package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	errNotELF = errors.New("target is not an ELF file")
)

func main() {
	var adb string
	var err error

	if adb, err = findAdb(); err != nil {
		printlnSilent(os.Stderr, "adb or adb.exe not found")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printlnSilent(os.Stderr, "executable or zip path not found")
		os.Exit(1)
	}

	targetFile := os.Args[1]
	otherArgs := os.Args[2:]

	if isZip(targetFile) {
		if len(otherArgs) == 0 {
			printlnSilent(os.Stderr, "main class name not found")
			os.Exit(1)
		}
		err = runDexJar(adb, targetFile, otherArgs)
	} else {
		extras, otherArgs := findExtraFiles(otherArgs)
		err = runExec(adb, targetFile, extras, otherArgs)
	}

	exit(err)
}

func findAdb() (string, error) {
	return exec.LookPath("adb")
}

func exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printlnSilent(w io.Writer, a ...interface{}) {
	if _, err := fmt.Fprintln(w, a...); err != nil {
		exit(err)
	}
}

func closeSilent(f *os.File) {
	if err := f.Close(); err != nil {
		exit(err)
	}
}

func findExtraFiles(args []string) ([]string, []string) {
	max := len(args)
	if max == 0 {
		return nil, args
	}

	var ret []string
	index := 0

	for ; index < max; index++ {
		if info, err := os.Stat(args[index]); err == nil && (info.Mode()&fs.ModeType) == 0 {
			ret = append(ret, args[index])
		} else {
			break
		}
	}

	return ret, args[index:]
}

func runExec(adb string, target string, extraFiles []string, oArgs []string) error {
	if !isELF(target) {
		return errNotELF
	}

	// adb push ...
	totalFiles := []string{target}
	totalFiles = append(totalFiles, extraFiles...)

	for _, f := range totalFiles {
		execFile, err := filepath.Abs(f)
		if err != nil {
			return err
		}
		fmt.Printf("prepare to push %s to device\n", execFile)
		if err := runCmd(adb, "push", execFile, "/data/local/tmp/"); err != nil {
			return err
		}
	}

	var args []string
	fileTargetPath := "/data/local/tmp/" + filepath.Base(target)
	args = append(args, "shell",
		"cd /data/local/tmp/",
		"&& chmod +x", fileTargetPath,
		"&& echo \"[program output]\" &&",
		"time",
		"sh", "-c",
		"'",
		"LD_LIBRARY_PATH=/data/local/tmp",
		fileTargetPath)
	args = append(args, oArgs...)
	args = append(args, "&& echo \"[program execution completed]\" || echo \"[error code returned: ($?)]\"", "'")

	for _, f := range totalFiles {
		args = append(args, fmt.Sprintf("&& rm /data/local/tmp/%s", filepath.Base(f)))
	}

	return runCmd(adb, args...)
}

func runDexJar(adb string, target string, oArgs []string) error {
	execFile, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	fmt.Printf("prepare to push %s to device\n", execFile)
	if err := runCmd(adb, "push", execFile, "/data/local/tmp/"); err != nil {
		return err
	}
	var args []string
	args = append(args, "shell",
		"cd /data/local/tmp/",
		"&& echo \"[program output]\" &&",
		"time",
		"sh", "-c",
		"'",
		"LD_LIBRARY_PATH=/data/local/tmp",
		fmt.Sprintf("CLASSPATH=\"/data/local/tmp/%s\"", filepath.Base(execFile)),
		"app_process",
		"/")
	args = append(args, oArgs...)
	args = append(args, "&& echo \"[program execution completed]\" || echo \"[error code returned: ($?)]\"",
		"'",
		"&& rm "+fmt.Sprintf("\"/data/local/tmp/%s\"", filepath.Base(execFile)))
	err = runCmd(adb, args...)
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
	defer closeSilent(f)
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
	defer closeSilent(f)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return false
	}
	return buf[0] == 0x7F && buf[1] == 0x45 && buf[2] == 0x4C && buf[3] == 0x46
}
