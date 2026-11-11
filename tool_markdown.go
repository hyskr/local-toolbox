package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"path/filepath"
)

func (a *app) mdExport(output string) (string, error) {
	if a.md == "" {
		return "", fmt.Errorf("请选择 Markdown 文件")
	}
	args := []string{a.md, "-o", output, "--margin", num(a.margin), "--line-height", num(a.line), "--paragraph-spacing", num(a.spacing), "--font-size", num(a.fontsize)}
	if a.compact {
		args = append(args, "--compact")
	}
	if err := run("liubai-pdf", args...); err != nil {
		return "", err
	}
	return output, nil
}
func (a *app) markdownKey() string {
	return fmt.Sprintf("%s|%t|%g|%g|%g|%g", a.md, a.compact, a.margin, a.line, a.spacing, a.fontsize)
}
func (a *app) pickMarkdown() {
	a.pickFile("选择 Markdown", []string{"md", "markdown", "txt"}, func(path string) { a.md = path; a.preview = nil; a.mdPreviewKey = "" })
}
func (a *app) previewMarkdown() {
	snapshot := *a
	a.job("生成预览", func() (string, error) {
		path := filepath.Join(a.temp, "markdown-preview.pdf")
		if _, err := snapshot.mdExport(path); err != nil {
			return "", err
		}
		if err := a.renderPreview(path); err != nil {
			return "", err
		}
		a.win.Update(func() { a.mdPreviewKey = snapshot.markdownKey() })
		return path, nil
	})
}
func (a *app) saveMarkdownPDF() {
	snapshot := *a
	name := fileStem(a.md) + ".pdf"
	a.save("导出 PDF", name, "pdf", snapshot.mdExport)
}
func fileStem(path string) string {
	if path == "" {
		return "文稿"
	}
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}
func (a *app) markdown(c *ui.Context) {
	pageHeader(c, "markdown", "Markdown 转 PDF", func() {
		if a.md != "" {
			if ui.Button(c, "选择文稿…").Disabled(a.working).Clicked() {
				a.pickMarkdown()
			}
			if ui.PrimaryButton(c, "导出 PDF…").Disabled(a.working).Clicked() {
				a.saveMarkdownPDF()
			}
		}
	})
	if a.md == "" {
		emptyStage(c, "markdown", "拖入 Markdown 文稿", "选择文稿…", a.pickMarkdown)
		return
	}
	workspace(c, func() {
		a.sourceFile(c, a.md, "源文件", a.pickMarkdown)
		ui.Divider(c)
		section(c, "排版", func() {
			preset := 0
			if a.compact {
				preset = 1
			}
			if ui.Segmented(c, &preset, "舒展", "紧凑").FillWidth().Disabled(a.working).Changed() {
				a.compact = preset == 1
				if a.compact {
					a.margin = 12
					a.line = 1.55
					a.spacing = 8
				} else {
					a.margin = 18
					a.line = 1.9
					a.spacing = 14
				}
			}
			numberSetting(c, "字号 · px", &a.fontsize, 8, 32, 1)
			numberSetting(c, "行距", &a.line, 1, 3, .05)
			numberSetting(c, "段落间距 · px", &a.spacing, 0, 80, 1)
			numberSetting(c, "页边距 · mm", &a.margin, 0, 80, 1)
		})
		if ui.Button(c, "恢复默认排版").Disabled(a.working).Clicked() {
			a.compact = false
			a.margin = 18
			a.line = 1.9
			a.spacing = 14
			a.fontsize = 16
		}
	}, func() {
		paneBar(c, "A4 · 第一页", func() {
			stale := a.preview != nil && a.mdPreviewKey != a.markdownKey()
			if stale {
				ui.Text(c, "设置已更改").FontSize(11).TextColor(c.Theme().TextMuted)
			}
			if ui.Button(c, "更新预览").Disabled(a.working).Clicked() {
				a.previewMarkdown()
			}
		})
		if a.preview == nil {
			emptyStage(c, "markdown", "预览排版效果", "生成预览", func() {
				if !a.working {
					a.previewMarkdown()
				}
			})
		} else {
			paperPreview(c, a.preview)
		}
	})
}
