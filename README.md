# arun

辅助 CLion 运行 Android 上可执行的 C/C++ 程序。

#### 初衷

每一次编写可以在 Android 设备上运行的 Hello world 程序的过程都是非常繁琐的：

* 使用 NDK 编译可执行程序
* adb push 到 Android 设备中
* adb shell 运行可执行程序

如此反复。

工具 arun 就是为了简化这个过程。

#### 安装

```bash
go get github.com/ClarkGuan/arun
```

#### 使用

```bash
arun -p <CLion 工程路径> -m <debug 或 release> <程序参数列表>
```

例如，在 CLion 工程目录下执行运行：

```bash
arun
```

或传递参数 x, y, z：

```bash
arun x y z
```

如果在其他目录则指定 CLion 的工程目录：

```bash
arun -p ../hello_world_project x y z
```

默认情况下，arun 识别 CLion 的 debug 编译产物，即 cmake-build-debug 目录，我们可以使用 `-m` 选项指定其他编译模式，例如

```bash
arun -m release
```

这样，arun 工具会寻找 cmake-build-release 目录中的可执行文件。
