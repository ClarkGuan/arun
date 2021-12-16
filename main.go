package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	errNotELFOrDex = errors.New("target is not an ELF of DEX file")
	errNoMainClass = errors.New("no main class found")
)

const ARunCopy = "ARUN_COPY"
const ARunVerbose = "ARUN_VERBOSE"

func main() {
	var adb string
	var err error

	if adb, err = findAdb(); err != nil {
		printlnSilent(os.Stderr, "adb or adb.exe not found")
		os.Exit(1)
	}

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose output")

	flag.Parse()

	v := strings.ToLower(strings.TrimSpace(os.Getenv(ARunVerbose)))
	verbose = verbose || v == "1" || v == "true"

	if len(flag.Args()) < 2 {
		printlnSilent(os.Stderr, "executable or zip path not found")
		os.Exit(1)
	}

	targetFile := flag.Arg(0)
	otherArgs := flag.Args()[1:]
	exit(NewRunnable().SetVerbose(verbose).SetExtras(findExtraFiles()).Run(adb, targetFile, otherArgs))
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

func regularFileExist(s string) bool {
	info, err := os.Stat(s)
	return err == nil && info.Mode()&fs.ModeType == 0
}

func findExtraFiles() []string {
	val, found := os.LookupEnv(ARunCopy)
	if !found {
		return nil
	}
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return nil
	}
	vars := strings.Split(val, ":")
	var ret []string
	for _, s := range vars {
		if regularFileExist(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

type RunnableType int

const (
	UNKNOWN RunnableType = iota
	EXEC
	DEX
)

type Runnable struct {
	typo    RunnableType
	extras  []string
	verbose bool
}

func NewRunnable() *Runnable {
	return &Runnable{typo: UNKNOWN}
}

func (runnable *Runnable) SetExtras(extras []string) *Runnable {
	runnable.extras = extras
	return runnable
}

func (runnable *Runnable) SetType(t RunnableType) *Runnable {
	switch t {
	case EXEC | DEX:
		runnable.typo = t
	default:
		runnable.typo = UNKNOWN
	}
	return runnable
}

func (runnable *Runnable) makeTypeExist(target string) error {
	if runnable.typo < EXEC || runnable.typo > DEX {
		if isELF(target) {
			runnable.SetType(EXEC)
		} else if isZip(target) {
			runnable.SetType(DEX)
		} else {
			return errNotELFOrDex
		}
	}

	return nil
}

func (runnable *Runnable) SetVerbose(b bool) *Runnable {
	runnable.verbose = b
	return runnable
}

func (runnable *Runnable) Run(adb, target string, oArgs []string) error {
	if err := runnable.makeTypeExist(target); err != nil {
		return err
	}

	if runnable.typo == DEX && len(oArgs) == 0 {
		return errNoMainClass
	}

	totalPushFiles := []string{target}
	totalPushFiles = append(totalPushFiles, runnable.extras...)
	var needRemoveFiles []string

	for _, path := range totalPushFiles {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if runnable.verbose {
			fmt.Printf("prepare to push %s to device\n", absPath)
		}
		if err = runCmd(adb, "push", absPath, "/data/local/tmp/"); err != nil {
			return err
		}
		needRemoveFiles = append(needRemoveFiles, fmt.Sprintf("/data/local/tmp/%s", filepath.Base(absPath)))
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
		"LD_LIBRARY_PATH=/data/local/tmp")

	switch runnable.typo {
	case EXEC:
		args = append(args, fileTargetPath)

	case DEX:
		args = append(args, fmt.Sprintf("CLASSPATH=\"%s\"", fileTargetPath),
			"app_process", "/")

	default:
		panic("never reach here")
	}

	args = append(args, oArgs...)
	args = append(args, "&& echo \"[program execution completed]\" || echo \"[error code returned: ($?)]\"", "'")

	for _, path := range needRemoveFiles {
		args = append(args, fmt.Sprintf("&& rm \"%s\"", path))
	}

	return runCmd(adb, args...)
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
