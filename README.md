# arun

辅助运行 Android 上可执行程序。

#### 初衷

每一次编写可以在 Android 设备上运行的 Hello world 程序的过程都是非常繁琐的：

* 使用 NDK 编译可执行程序
* adb push 到 Android 设备中
* adb shell 运行可执行程序

如此反复。

工具 arun 就是为了简化这个过程。

#### 安装

```bash
GO111MODULE=off go get github.com/ClarkGuan/arun
```

或

```bash
git clone https://github.com/ClarkGuan/arun
cd arun
go install
```

#### 使用

```bash
arun [-exe|exe] <可执行文件路径> <程序参数列表>
```

另外也增加了对于 Golang 编译的单元测试二进制文件的支持：

```bash
arun [-test|test] <Go单元测试可执行文件路径> <程序参数列表>
```

具体可以参考我编写的另一个工具 —— ddmobile（https://github.com/ClarkGuan/ddmobile）。
