# di

`di` 是一个可断开、可重新进入的终端会话工具。它用 Go 自己管理 PTY 和 Unix socket，不依赖 `dtach`。

## 依赖

默认使用内建的纯文本选择器，不依赖 `fzf`。

如果你想继续使用 `fzf`，可以显式指定：

```sh
DI_PICKER=fzf di
```

启动新会话和列出会话同样不依赖 `fzf`：

```sh
d <command> [args...]
d --list
```

## 安装

```sh
git clone git@github.com:whoamihappyhacking/di.git
cd di
go build -o di .
./di install
```

安装后：

```text
~/.local/bin/d
~/.local/bin/di -> ~/.local/bin/d
```

确保 `~/.local/bin` 在 `PATH` 里。

## 用法

查看命令说明：

```sh
d --help
di --help
```

启动一个会话：

```sh
d codex --yolo
```

断开 attach，后端命令继续运行：

```text
Ctrl-]
```

鼠标滚轮不会转发给后端程序，方便用终端自己的滚屏查看历史输出。

选择已有会话：

```sh
di
```

`di` 默认会显示一个编号列表，让你选择 session。即使在同一个目录里重复执行同一个命令，也会创建不同的 session。

### TUI 模式

`di tui` 提供一个终端界面，左侧显示 session 列表，右侧实时预览选中 session 的输出（每 500ms 刷新）。

```sh
di tui
```

快捷键：

| 按键 | 功能 |
|------|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `g` | 跳到第一个 session |
| `G` | 跳到最后一个 session |
| `Enter` | 进入选中的 session |
| `q` | 退出 TUI |

Enter 进入 session 后，TUI 退出，`attach()` 接管终端。Ctrl-] 断开后回到 shell，需要重新运行 `di tui` 才能再次浏览。

注意：TUI 预览的是 PTY 原始输出（含 ANSI 转义），对 shell/CLI 程序效果较好，全屏应用（如 vim）为 best-effort。

如果想用 `fzf` 选择器：

```sh
DI_PICKER=fzf di
```

列出会话：

```sh
d --list
```

从另一个终端断开某个 attach 客户端：

```sh
d --detach codex---yolo
```

临时修改 detach 快捷键：

```sh
D_DETACH='^B' di
```

## 构建

Linux/macOS 都支持。

```sh
go build -o di .
GOOS=darwin GOARCH=arm64 go build -o di-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o di-darwin-amd64 .
```

## 说明

`di` 解决的是”终端断开后重新进入”的问题，不是 checkpoint 工具；它不会保存进程内存、文件系统快照或网络连接状态。
