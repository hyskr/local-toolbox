package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

var conversionFormats = map[string][]string{
	"图片":  {"PNG", "JPG", "HEIC", "WEBP", "TIFF", "GIF", "BMP", "PDF"},
	"PDF": {"PNG", "JPG", "TIFF", "HEIC", "TXT", "MD", "DOCX", "RTF"},
	"文档":  {"DOCX", "PDF", "RTF", "MD", "TXT", "HTML", "ODT", "DOC"},
	"幻灯片": {"PDF", "MD"},
	"表格":  {"PDF", "MD"},
	"视频":  {"MP4", "MOV", "GIF", "M4A"},
	"音频":  {"M4A", "WAV", "AIFF", "CAF"},
}

func fileConversionKind(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch ext {
	case "png", "jpg", "jpeg", "jpe", "heic", "heif", "webp", "tiff", "tif", "gif", "bmp", "dng", "raw", "arw", "cr2", "cr3", "crw", "nef", "nrw", "orf", "rw2", "raf", "pef", "srw", "3fr", "erf", "iiq", "k25", "kdc", "mos", "mrw", "rwl", "x3f":
		return "图片"
	case "pdf":
		return "PDF"
	case "docx", "doc", "rtf", "rtfd", "odt", "html", "htm", "txt", "text":
		return "文档"
	case "pptx":
		return "幻灯片"
	case "xlsx":
		return "表格"
	case "mov", "mp4", "m4v":
		return "视频"
	case "m4a", "mp3", "wav", "aiff", "aif", "caf", "aac":
		return "音频"
	default:
		return ""
	}
}
func (a *app) converterScript() (string, error) {
	files := []string{"Support.swift", "Palette.swift", "BackgroundRemover.swift", "DocumentConverter.swift", "ImageConverter.swift", "MarkdownWriter.swift", "MediaConverter.swift", "PDFConverter.swift", "PresentationConverter.swift", "SpreadsheetConverter.swift", "TextRecognizer.swift", "WordAutomation.swift", "WordMarkdown.swift", "WordReader.swift", "ZipArchive.swift", "Runner.swift"}
	var source strings.Builder
	for _, name := range files {
		b, err := scripts.ReadFile("scripts/fileflipper/" + name)
		if err != nil {
			return "", err
		}
		source.Write(b)
		source.WriteByte('\n')
	}
	path := filepath.Join(a.temp, "fileflipper.swift")
	if err := os.WriteFile(path, []byte(source.String()), 0600); err != nil {
		return "", err
	}
	return path, nil
}
func (a *app) chooseConversionFiles() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: "选择文件", Multiple: true})
		a.win.Update(func() {
			if err != nil {
				a.status = err.Error()
			} else {
				a.acceptConversionFiles(paths)
			}
		})
	}()
}
func (a *app) showConversionResult(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		a.status = err.Error()
		return
	}
	a.conversionText = string(data)
	a.conversionOutput = path
}
func (a *app) saveConversionText() {
	text := a.conversionText
	name := filepath.Base(a.conversionOutput)
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	a.save("保存结果", name, ext, func(out string) (string, error) { return out, os.WriteFile(out, []byte(text), 0644) })
}
func (a *app) convertPage(c *ui.Context) {
	pageHeader(c, "convert", "格式转换", func() {
		if len(a.conversionFiles) > 0 {
			if ui.Button(c, "选择文件…").Disabled(a.working).Clicked() {
				a.chooseConversionFiles()
			}
			if a.conversionCancel != nil {
				if ui.Button(c, "停止任务").Clicked() {
					a.conversionCancel()
				}
			} else if ui.PrimaryButton(c, "开始转换").Disabled(a.working).Clicked() {
				a.startConversion()
			}
		}
	})
	if len(a.conversionFiles) == 0 {
		emptyStage(c, "convert", "拖入文件，选择目标格式", "选择文件…", a.chooseConversionFiles)
		return
	}
	workspace(c, func() {
		section(c, "本次任务", func() {
			ui.Text(c, fmt.Sprintf("%d 个%s文件", len(a.conversionFiles), a.conversionKind)).FontSize(14).Bold()
			if ui.Button(c, "清空列表").Disabled(a.working).Clicked() {
				a.conversionFiles = nil
				a.conversionItems = nil
				a.conversionOutput = ""
				a.conversionText = ""
				a.conversionKind = ""
				a.conversionFormat = ""
			}
		})
		ui.Divider(c)
		section(c, "目标格式", func() {
			ui.Select(c, &a.conversionFormat, availableFormats(a.conversionKind, a.conversionFiles)).FillWidth().Disabled(a.working).Label("目标格式")
		})
		section(c, "保存到", func() {
			label := "原文件旁"
			if a.conversionFolder != "" {
				label = a.conversionFolder
			}
			compactPath(c, label)
			ui.Row(c).Gap(6).Children(func() {
				if ui.Button(c, "选择文件夹…").Disabled(a.working).Clicked() {
					a.pickDir(func(path string) { a.conversionFolder = path })
				}
				if a.conversionFolder != "" && iconButton(c, "close", "恢复保存到原文件旁", a.working).Clicked() {
					a.conversionFolder = ""
				}
			})
		})
		if a.conversionKind == "图片" {
			ui.Divider(c)
			if ui.Button(c, "识别这些图片的文字").FillWidth().Disabled(a.working).Clicked() {
				a.openOCR(a.conversionFiles)
			}
		}
		if a.conversionCancel == nil && hasFailedConversions(a.conversionItems) {
			ui.Divider(c)
			if ui.Button(c, "重试失败项").FillWidth().Disabled(a.working).Clicked() {
				paths := []string{}
				for _, item := range a.conversionItems {
					if item.State == "failed" {
						paths = append(paths, item.Path)
					}
				}
				a.startConversionPaths(paths, true)
			}
		}
	}, func() {
		if a.conversionOutput != "" {
			a.conversionResultView(c)
		} else {
			a.conversionListView(c)
		}
	})
}
func (a *app) conversionListView(c *ui.Context) {
	done, failed, finished := 0, 0, 0
	for _, item := range a.conversionItems {
		switch item.State {
		case "done":
			done++
			finished++
		case "failed":
			failed++
			finished++
		case "cancelled":
			finished++
		}
	}
	paneBar(c, "任务列表", func() {
		if len(a.conversionItems) > 0 {
			ui.Text(c, fmt.Sprintf("%d 完成 · %d 失败", done, failed)).FontSize(12).TextColor(c.Theme().TextMuted)
		} else {
			ui.Text(c, "等待转换").FontSize(12).TextColor(c.Theme().TextMuted)
		}
	})
	if len(a.conversionItems) > 0 {
		ui.Progress(c, float64(finished)/float64(len(a.conversionItems))).Label("批量转换进度")
	}
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
		ui.Column(c).FillWidth().Padding(16, 18).Gap(8).Children(func() {
			if len(a.conversionItems) == 0 {
				for i, path := range a.conversionFiles {
					i, path := i, path
					a.conversionRow(c, conversionItem{Path: path, State: "pending"}, func() { a.conversionFiles = append(a.conversionFiles[:i], a.conversionFiles[i+1:]...) })
				}
			} else {
				for _, item := range a.conversionItems {
					a.conversionRow(c, item, nil)
				}
			}
		})
	})
}
func (a *app) conversionRow(c *ui.Context, item conversionItem, remove func()) {
	ui.Column(c).FillWidth().Padding(12, 14).Gap(8).Radius(8).Background(colors(c).surface).Children(func() {
		ui.Row(c).FillWidth().Gap(10).Children(func() {
			ui.Icon(c, toolIcons["markdown"]).Size(18, 18).TextColor(c.Theme().TextMuted)
			ui.Text(c, filepath.Base(item.Path)).Grow(1).MinWidth(0).FontSize(13).SingleLine().Tooltip(item.Path)
			col := c.Theme().TextMuted
			if item.State == "done" {
				col = c.Theme().Success
			} else if item.State == "failed" {
				col = c.Theme().Danger
			}
			ui.Text(c, conversionStateLabel(item.State)).FontSize(11).TextColor(col)
			if iconButton(c, "folder", "打开源文件夹", false).Clicked() {
				a.openSourceFolder(item.Path)
			}
			if remove != nil {
				iconButton(c, "close", "移除文件", a.working).OnClick(remove)
			}
		})
		if item.Error != "" {
			ui.Text(c, item.Error).FontSize(12).TextColor(c.Theme().Danger).Wrap()
		}
		for _, output := range item.Outputs {
			output := output
			ui.Row(c).FillWidth().Gap(8).Children(func() {
				ui.Text(c, filepath.Base(output)).Grow(1).MinWidth(0).FontSize(12).TextColor(c.Theme().TextMuted).SingleLine()
				if isTextResult(output) && ui.Button(c, "查看结果").Clicked() {
					a.showConversionResult(output)
				}
				if iconButton(c, "open", "打开结果", false).Clicked() {
					_ = exec.Command("open", output).Start()
				}
			})
		}
	})
}
func (a *app) conversionResultView(c *ui.Context) {
	ui.Row(c).FillWidth().Height(50).Shrink(0).Padding(0, 16).Gap(8).Children(func() {
		iconButton(c, "back", "返回任务列表", false).OnClick(func() { a.conversionOutput = "" })
		ui.Text(c, filepath.Base(a.conversionOutput)).Grow(1).MinWidth(0).FontSize(12).Bold().SingleLine()
		if ui.Button(c, "复制").Clicked() {
			mygo.Clipboard.WriteText(a.conversionText)
			a.status = "已复制"
		}
		if ui.Button(c, "另存为…").Disabled(a.working).Clicked() {
			a.saveConversionText()
		}
	})
	ui.Column(c).Grow(1).MinHeight(0).FillWidth().Padding(0, 18, 18, 18).Children(func() { ui.TextArea(c, &a.conversionText).Grow(1).MinHeight(0).FillWidth().Label("转换结果") })
}

func availableFormats(kind string, paths []string) []string {
	formats := conversionFormats[kind]
	out := []string{}
	for _, format := range formats {
		allSame := true
		for _, path := range paths {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
			switch ext {
			case "jpeg", "jpe":
				ext = "jpg"
			case "tif":
				ext = "tiff"
			case "heif":
				ext = "heic"
			case "aif":
				ext = "aiff"
			}
			if ext != strings.ToLower(format) {
				allSame = false
				break
			}
		}
		if !allSame {
			out = append(out, format)
		}
	}
	return out
}

func (a *app) acceptConversionFiles(paths []string) {
	if a.working || len(paths) == 0 {
		return
	}
	kind := fileConversionKind(paths[0])
	if kind == "" {
		a.status = "不支持这个文件类型"
		return
	}
	for _, path := range paths {
		if fileConversionKind(path) != kind {
			a.status = "请选择同一类型的文件"
			return
		}
	}
	a.conversionItems = nil
	a.conversionOutput = ""
	a.conversionText = ""
	a.conversionFiles = append([]string(nil), paths...)
	a.conversionKind = kind
	options := availableFormats(kind, paths)
	a.conversionFormat = options[0]
	for _, option := range options {
		if option == a.formatByKind[kind] {
			a.conversionFormat = option
			break
		}
	}
	a.status = ""
	a.last = ""
}
