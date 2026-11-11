package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func (a *app) recognizeImages(paths []string) (string, error) {
	script, err := a.script("ocr.swift")
	if err != nil {
		return "", err
	}
	result, err := runOutput("swift", append([]string{"-suppress-warnings", script}, paths...)...)
	if err != nil {
		return "", err
	}
	var texts []string
	if err = json.Unmarshal([]byte(result), &texts); err != nil {
		return "", err
	}
	if len(texts) != len(paths) {
		return "", fmt.Errorf("识别结果数量不一致")
	}
	var parts []string
	for i, text := range texts {
		if len(paths) > 1 {
			parts = append(parts, filepath.Base(paths[i])+"\n"+text)
		} else {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}
func (a *app) selectOCR(index int) {
	a.ocrSelection = index
	a.ocrBitmap = nil
	if index < 0 || index >= len(a.ocrFiles) {
		return
	}
	data, err := os.ReadFile(a.ocrFiles[index])
	if err == nil {
		a.ocrBitmap, _ = ui.DecodeBitmap(data)
	}
}
func (a *app) pickOCR() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: "选择图片", Multiple: true, Filters: []mygo.FileFilter{{Name: "图片", Extensions: []string{"png", "jpg", "jpeg", "heic", "heif", "tif", "tiff", "bmp", "webp"}}}})
		a.win.Update(func() {
			if err != nil {
				a.status = err.Error()
			} else if len(paths) > 0 && !a.ocrBusy {
				a.ocrFiles = paths
				a.ocrSources = append([]string(nil), paths...)
				a.ocrText = ""
				a.status = ""
				a.selectOCR(0)
			}
		})
	}()
}
func (a *app) startOCR() {
	if a.ocrBusy || len(a.ocrFiles) == 0 {
		return
	}
	paths := append([]string(nil), a.ocrFiles...)
	owner := a.page
	a.ocrBusy = true
	a.status = "识别图片…"
	a.last = ""
	go func() {
		text, err := a.recognizeImages(paths)
		a.win.Update(func() {
			a.ocrBusy = false
			if err != nil {
				a.setNotice(owner, err.Error(), "")
			} else {
				a.ocrText = text
				message := "识别完成"
				if strings.TrimSpace(text) == "" {
					message = "没有识别到文字"
				}
				a.setNotice(owner, message, "")
			}
		})
	}()
}
func (a *app) saveOCR() {
	text := a.ocrText
	name := "识别结果.txt"
	if len(a.ocrSources) == 1 {
		name = fileStem(a.ocrSources[0]) + ".txt"
	}
	a.save("保存文字", name, "txt", func(out string) (string, error) { return out, os.WriteFile(out, []byte(text), 0644) })
}
func (a *app) ocrPage(c *ui.Context) {
	pageHeader(c, "ocr", "图片文字识别", func() {
		if len(a.ocrFiles) > 0 {
			if ui.Button(c, "选择图片…").Disabled(a.ocrBusy).Clicked() {
				a.pickOCR()
			}
			label := "识别文字"
			if a.ocrBusy {
				label = "识别中…"
			}
			if ui.PrimaryButton(c, label).Disabled(a.ocrBusy).Clicked() {
				a.startOCR()
			}
		}
	})
	if len(a.ocrFiles) == 0 {
		emptyStage(c, "ocr", "拖入图片，提取文字", "选择图片…", a.pickOCR)
		return
	}
	workspace(c, func() {
		section(c, fmt.Sprintf("图片 · %d 张", len(a.ocrFiles)), func() {
			for i, path := range a.ocrFiles {
				i, path := i, path
				name := path
				if i < len(a.ocrSources) {
					name = a.ocrSources[i]
				}
				row := ui.ButtonBase(c).FillWidth().Height(38).Padding(0, 8).Gap(8).Justify(ui.Start).Radius(6).Disabled(a.ocrBusy).Label(filepath.Base(name))
				if i == a.ocrSelection {
					row.Background(colors(c).selected)
				} else if row.Hovered() {
					row.Background(c.Theme().SurfaceHover)
				}
				row.Children(func() {
					ui.Icon(c, toolIcons["photos"]).Size(16, 16).TextColor(c.Theme().TextMuted)
					ui.Text(c, filepath.Base(name)).Grow(1).MinWidth(0).SingleLine().FontSize(12)
				})
				row.OnClick(func() { a.selectOCR(i) })
			}
		})
		if a.ocrSelection < len(a.ocrFiles) {
			i := a.ocrSelection
			name := a.ocrFiles[i]
			if i < len(a.ocrSources) {
				name = a.ocrSources[i]
			}
			ui.Row(c).Gap(4).Children(func() {
				if iconButton(c, "folder", "打开源文件夹", false).Clicked() {
					a.openSourceFolder(name)
				}
				if iconButton(c, "close", "移除图片", a.ocrBusy).Clicked() {
					a.ocrFiles = append(a.ocrFiles[:i], a.ocrFiles[i+1:]...)
					if i < len(a.ocrSources) {
						a.ocrSources = append(a.ocrSources[:i], a.ocrSources[i+1:]...)
					}
					a.ocrText = ""
					a.selectOCR(0)
				}
			})
			if a.ocrBitmap != nil {
				ui.Divider(c)
				ui.Image(c, a.ocrBitmap).FillWidth().Height(180).Fit(ui.Contain)
			}
		}
	}, func() {
		paneBar(c, "识别结果", func() {
			if ui.Button(c, "复制").Disabled(a.ocrText == "" || a.ocrBusy).Clicked() {
				mygo.Clipboard.WriteText(a.ocrText)
				a.status = "已复制"
			}
			if ui.Button(c, "保存 TXT…").Disabled(a.ocrText == "" || a.ocrBusy || a.working).Clicked() {
				a.saveOCR()
			}
		})
		if a.ocrBusy {
			ui.Progress(c, -1).Label("正在识别图片")
		}
		ui.Column(c).Grow(1).MinHeight(0).FillWidth().Padding(0, 18, 18, 18).Children(func() {
			ui.TextArea(c, &a.ocrText).Grow(1).MinHeight(0).FillWidth().Placeholder("识别后可在这里编辑文字").Label("识别结果").Disabled(a.ocrBusy)
		})
	})
}
func (a *app) openOCR(paths []string) {
	for i, tool := range tools {
		if tool.id == "ocr" {
			if !a.enabled[tool.id] {
				a.status = "请在插件管理中启用图片文字识别"
				return
			}
			if a.ocrBusy {
				a.status = "文字识别正在运行"
				return
			}
			a.ocrFiles = append([]string(nil), paths...)
			a.ocrSources = append([]string(nil), paths...)
			for i, path := range paths {
				for _, photo := range a.photos {
					if photo.Path == path {
						a.ocrSources[i] = photo.Original
						break
					}
				}
			}
			a.ocrText = ""
			a.selectOCR(0)
			a.navigate(i)
			return
		}
	}
}
