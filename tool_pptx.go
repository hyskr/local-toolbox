package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
)

func (a *app) pickPresentationPDF() { a.pickFile("选择 PDF", []string{"pdf"}, a.setPresentationPDF) }
func (a *app) setPresentationPDF(path string) {
	a.pdf = path
	a.pptPreview = nil
	a.pptPreviewPath = path
	a.pptPreviewBusy = true
	go func() {
		image, err := a.pdfImage(path)
		a.win.Update(func() {
			if a.pptPreviewPath == path {
				a.pptPreviewBusy = false
				if err == nil {
					a.pptPreview = image
				}
			}
		})
	}()
}
func (a *app) pdfImage(path string) (*ui.Bitmap, error) {
	dir, err := os.MkdirTemp(a.temp, "pdf-preview-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	prefix := filepath.Join(dir, "page")
	if err = run("pdftoppm", "-f", "1", "-l", "1", "-singlefile", "-scale-to", "1000", "-png", path, prefix); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(prefix + ".png")
	if err != nil {
		return nil, err
	}
	return ui.DecodeBitmap(data)
}
func (a *app) pptxReady() bool {
	return a.notesKind == "Typst 源文件" && a.source != "" || a.pdf != "" && (a.notesKind == "不使用备注" || a.notesKind == "pdfpc 备注数据" && a.notes != "")
}
func (a *app) exportPPTX() {
	pdf, source, notes, resolution := a.pdf, a.source, a.notes, a.resolution
	if a.notesKind != "Typst 源文件" {
		source = ""
	}
	if a.notesKind != "pdfpc 备注数据" {
		notes = ""
	}
	name := fileStem(pdf)
	if source != "" {
		name = fileStem(source)
	}
	a.save("导出 PPTX", name+".pptx", "pptx", func(out string) (string, error) {
		if pdf == "" && source == "" {
			return "", fmt.Errorf("请选择输入文件")
		}
		script, err := a.script("pdf2pptx.py")
		if err != nil {
			return "", err
		}
		pdfPath := pdf
		if source != "" {
			pdfPath = filepath.Join(a.temp, "typst-export.pdf")
		}
		args := []string{script, pdfPath, out, "--width", num(resolution)}
		if source != "" {
			args = append(args, "--source", source)
			fonts := filepath.Join(filepath.Dir(source), "fonts")
			if _, err := os.Stat(fonts); err == nil {
				args = append(args, "--font-path", fonts)
			}
		}
		if notes != "" {
			args = append(args, "--notes", notes)
		}
		if err = run("python3", args...); err != nil {
			return "", err
		}
		return out, nil
	})
}
func (a *app) pptx(c *ui.Context) {
	pageHeader(c, "pptx", "PDF 转 PPTX", func() {
		if ui.PrimaryButton(c, "导出 PPTX…").Disabled(!a.pptxReady() || a.working).Clicked() {
			a.exportPPTX()
		}
	})
	workspace(c, func() {
		section(c, "讲者备注", func() {
			ui.Select(c, &a.notesKind, []string{"不使用备注", "pdfpc 备注数据", "Typst 源文件"}).FillWidth().Disabled(a.working).Label("讲者备注来源")
		})
		if a.notesKind == "Typst 源文件" {
			a.sourceFile(c, a.source, "Typst 源文件", func() { a.pickFile("选择 Typst", []string{"typ"}, func(p string) { a.source = p }) })
		} else {
			a.sourceFile(c, a.pdf, "PDF 文件", a.pickPresentationPDF)
		}
		if a.notesKind == "pdfpc 备注数据" {
			a.sourceFile(c, a.notes, "讲者备注", func() { a.pickFile("选择备注", []string{"pdfpc", "json"}, func(p string) { a.notes = p }) })
		}
		ui.Divider(c)
		section(c, "输出", func() {
			numberSetting(c, "最长边 · px", &a.resolution, 640, 3840, 160)
			ui.Text(c, "每页作为一张图片，文字不可单独编辑。").FontSize(12).TextColor(c.Theme().TextMuted).Wrap()
		})
	}, func() {
		paneBar(c, "输入预览", nil)
		if a.notesKind == "Typst 源文件" {
			title := "选择 Typst 源文件"
			if a.source != "" {
				title = filepath.Base(a.source)
			}
			emptyStage(c, "templates", title, "", nil)
		} else if a.pptPreviewBusy {
			ui.Column(c).Grow(1).FillWidth().Center().Gap(12).Children(func() {
				ui.Spinner(c).Label("加载 PDF 预览")
				ui.Text(c, "加载预览…").FontSize(12).TextColor(c.Theme().TextMuted)
			})
		} else if a.pptPreview != nil {
			paperPreview(c, a.pptPreview)
		} else {
			title := "拖入 PDF 文件"
			button := "选择 PDF…"
			if a.pdf != "" {
				title = filepath.Base(a.pdf)
				button = "更换 PDF…"
			}
			emptyStage(c, "pptx", title, button, a.pickPresentationPDF)
		}
	})
}
