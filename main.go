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
	var exefile string
	var ddmobile string

	flag.StringVar(&mode, "m", "debug", "debug, release or relwithdebinfo...")
	flag.StringVar(&path, "clion", ".", "project path")
	flag.StringVar(&exefile, "exe", "", "executable file path")
	flag.StringVar(&ddmobile, "ddmobile", "", "ddmobile project path")
	flag.Parse()

	if len(exefile) != 0 {
		runSigleExeFile(exefile)
		return
	}

	if len(path) > 0 || len(mode) > 0 {
		runClionProject(mode, path)
		return
	}

	if len(ddmobile) > 0 {
		runDdmobileProject(ddmobile)
		return
	}

	if _, err := os.Stat("build/android/app"); err == nil {
		runDdmobileProject(".")
	}

	if _, err := os.Stat("cmake-build-debug"); err == nil {
		runClionProject("debug", ".")
	}

	flag.PrintDefaults()
}

func runSigleExeFile(exefile string) {
	execFile, err := filepath.Abs(exefile)
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
		"echo \"[程序输出如下]\" && LD_LIBRARY_PATH=/data/local/tmp",
		"/data/local/tmp/"+filepath.Base(execFile))
	args = append(args, flag.Args()...)
	args = append(args, "&& echo \"[程序执行完毕]\" || echo \"[程序执行返回$?]\"")
	if err := runCmd("adb", args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runDdmobileProject(path string) {
	buildDir := filepath.Join(path, "build/android/app")
	if infos, err := ioutil.ReadDir(buildDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		var files []string

		for _, info := range infos {
			if info.IsDir() {
				files = append(files, info.Name())
			}
		}

		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "can't find any ddmobile build files in %s", buildDir)
			os.Exit(1)
		}

		index := -1
		for index == -1 {
			for n, s := range files {
				fmt.Printf("%d: %s\n", n, s)
			}
			fmt.Printf("please enter your choice: ")
			fmt.Scanf("%d", &index)
			if index >= len(files) {
				fmt.Fprintf(os.Stderr, "index out of bounds")
				index = -1
			}
		}

		parentDir := filepath.Join(buildDir, files[index])
		if infos, err = ioutil.ReadDir(parentDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			target := filepath.Join(parentDir, infos[0].Name())
			runSigleExeFile(target)
		}
	}
}

func runClionProject(mode, path string) {
	cmakeBuildDir := filepath.Join(path, "cmake-build-"+mode)
	if _, err := os.Stat(cmakeBuildDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s not found\n", cmakeBuildDir)
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmakeBuildFile := filepath.Join(path, "CMakeLists.txt")
	targets := make([][]byte, 0, 8)
	if content, err := ioutil.ReadFile(cmakeBuildFile); err != nil {
		fmt.Fprintf(os.Stderr, "%s open fail", cmakeBuildFile)
		flag.PrintDefaults()
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
		flag.PrintDefaults()
		os.Exit(1)
	}

	runSigleExeFile(filepath.Join(cmakeBuildDir, string(targets[0])))
}

func runCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	return command.Run()
}
