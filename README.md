# 工具箱

使用 MyGo 原生 UI 的 macOS 本地工具箱，包含 Markdown 转 PDF、照片 A4 排版与扫描、PDF 转 PPTX、Typst 模板和文件格式转换。界面由 MyGo 绘制，不使用 WebView。

## 下载安装

从 [GitHub Releases](https://github.com/hyskr/local-toolbox/releases) 下载 macOS 通用版 DMG，打开后将「工具箱」拖入「应用程序」。应用支持 macOS 14 及更新版本，兼容 Apple 芯片与 Intel Mac。

当前版本使用临时签名，尚未经过 Apple 公证；首次打开时可在「系统设置 → 隐私与安全性」中确认打开。

## 运行依赖

- mise 管理项目 Go 1.27.4、Node.js 22.23.2 和 Python 3.14.7。
- [liubai-pdf](https://github.com/hyskr/liubai-pdf)、Node.js 及 Chrome/Chromium/Edge：Markdown 排版和导出。
- Python 3，安装 `Pillow`、`numpy`、`opencv-python`：照片扫描和 PDF 导出。
- Poppler（`pdfinfo`、`pdftoppm`）：PDF 预览和转 PPTX。
- macOS 14 或更新版本、Swift（Xcode 命令行工具）：文件格式转换。
- Typst 0.15.1 或更新版本，以及 `~/Documents/typst-templates` 模板库：模板项目创建。

各工具按需查找本机依赖。首次设置可以按需要安装：

```sh
brew install poppler typst
npm install -g github:hyskr/liubai-pdf
python3 -m pip install Pillow numpy opencv-python
git clone https://github.com/hyskr/typst-templates.git ~/Documents/typst-templates
```

图片格式转换和文字识别使用 macOS 系统框架及 Swift。Markdown 转 PDF 还需要安装 Chrome、Chromium 或 Edge。

## 从源码开发

开发时运行：

```sh
mise install
mise run dev
```

保存 Go 文件或项目配置后，MyGo 会重新编译并重启原生窗口；按 Ctrl+C 或关闭开发窗口停止。开发窗口与正式版使用独立的应用标识。打包应用使用 `mise run build`。

从 Finder 启动时也会查找 `~/.local/share/mise/shims`、`/opt/homebrew/bin` 和 `/usr/local/bin`。需要单独使用的工具缺少依赖时，界面会显示对应错误。

## 使用

首页按「转换与导出」和「图片与写作」选择工具，也可直接打开或拖入文件。工具页使用固定顶部操作栏、左侧设置区和右侧内容区；文件列表与预览分别滚动。

- Markdown：左侧调整排版，右侧更新第一页预览，顶部导出 PDF。
- 照片：在左侧选择及排序照片，右侧切换「排版预览」和「扫描校正」。扫描时依次点击左上、右上、右下、左下角，设置效果后应用；顶部导出 A4 PDF。
- PDF 转 PPTX：选择 PDF 或 Typst 输入、备注来源及图片尺寸，顶部导出。PDF 自动显示第一页预览；PPTX 页面为图片，文字不可单独编辑。
- Typst：左侧选择模板、项目名称与保存位置，右侧查看模板预览，顶部创建项目。
- 图片文字识别：左侧选择图片，顶部识别，右侧编辑、复制和保存文字。
- 格式转换：左侧选择格式及输出位置，右侧查看逐文件进度与结果；文字结果可查看、复制、另存，并可返回任务列表。支持停止及重试失败项。PPTX 转 PDF 时，图表、SmartArt 和动画不会绘制。

源文件旁的文件夹按钮可打开原文件所在位置。任务完成后的结果入口会随工具保留；切换页面后可返回继续查看。

快捷键：`⌘O` 选择文件，`⌘S` 导出或保存，`⌘,` 打开插件管理。相同命令位于原生菜单栏。

## 偏好与进度

排版参数、照片设置、模板及其保存位置、输出目录及各类文件的目标格式会保存在本机；不会保存选择的文件或识别内容。批量转换显示完成数量和逐文件状态，其他耗时任务显示处理中。

## 插件管理

工具作为内置模块注册，各工具的界面位于独立的 `tool_*.go` 文件。在「插件管理」中启用或停用，设置即时生效，并保存到用户配置目录的 `local-toolbox/plugins.json`。停用的工具不会出现在首页和侧栏。工具的任务在用户操作时才执行。

PDF 转 PPTX 脚本来自本机 typst-templates 项目。Markdown 导出调用 liubai-pdf；照片处理使用 OpenCV 与 Pillow。

文件格式转换使用 [FileFlipper](https://github.com/Aimee51819/FileFlipper) 的 macOS 转换代码；许可见 [FileFlipper 许可证](scripts/fileflipper/LICENSE)。

## 自动构建与发布

推送到 `main` 或提交 Pull Request 时，GitHub Actions 运行测试并生成 macOS 通用版 DMG。推送与 `mygo.json` 版本对应的 `v` 标签时，Actions 会创建 GitHub Release 并上传 DMG。
