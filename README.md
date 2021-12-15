# arun

辅助运行 Android 上可执行程序（可以是 C/C++/Go/Rust 等语言生成的）。

或者运行包含 classes.dex 文件的 jar/zip/apk 等压缩文件。

#### 初衷

每一次编写可以在 Android 设备上运行的 C/C++ 的 Hello world 程序的过程都是非常繁琐的：

* 使用 NDK 编译可执行程序
* adb push 到 Android 设备中
* adb shell 运行可执行程序

如此反复。

工具 arun 就是为了简化这个过程。

#### 安装

```bash
go install github.com/ClarkGuan/arun@latest
```

或

```bash
git clone https://github.com/ClarkGuan/arun && cd arun && go install
```

#### 使用

```bash
$ arun <可执行文件> <程序参数列表>
```

或

```bash
$ arun <可执行文件> <动态库文件列表> <程序参数列表>
```

或

```bash
$ arun <jar,zip,apk 文件> <Java主类> <程序参数列表>
```
