package main

import (
	"github.com/egoist/mygo"
	"strings"
)

func (a *app) openCurrent() {
	if a.working || a.ocrBusy {
		return
	}
	if a.page < 0 {
		go func() {
			paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: "打开文件", Multiple: true})
			a.win.Update(func() {
				if err != nil {
					a.status = err.Error()
				} else {
					a.acceptDrop(paths)
				}
			})
		}()
		return
	}
	switch tools[a.page].id {
	case "markdown":
		a.pickMarkdown()
	case "photos":
		a.pickPhotos()
	case "pptx":
		a.pickPresentationPDF()
	case "ocr":
		a.pickOCR()
	case "convert":
		a.chooseConversionFiles()
	case "templates":
		a.pickDir(func(p string) { a.folder = p })
	}
}
func (a *app) exportCurrent() {
	if a.working || a.ocrBusy {
		return
	}
	ready := false
	if a.page >= 0 && a.page < len(tools) {
		switch tools[a.page].id {
		case "markdown":
			ready = a.md != ""
			if ready {
				a.saveMarkdownPDF()
			}
		case "photos":
			ready = len(a.photos) > 0
			if ready {
				a.exportPhotos()
			}
		case "pptx":
			ready = a.pptxReady()
			if ready {
				a.exportPPTX()
			}
		case "ocr":
			ready = a.ocrText != ""
			if ready {
				a.saveOCR()
			}
		case "convert":
			ready = a.conversionOutput != ""
			if ready {
				a.saveConversionText()
			}
		case "templates":
			ready = a.folder != "" && strings.TrimSpace(a.project) != ""
			if ready {
				a.createTemplate()
			}
		}
	}
	if !ready {
		a.status = "请先选择文件并完成必要设置"
	}
}
func (a *app) installMenu() {
	invoke := func(action func()) func(*mygo.MenuItem, *mygo.Window) {
		return func(*mygo.MenuItem, *mygo.Window) { a.win.Update(action) }
	}
	toolMenu := []*mygo.MenuItem{{Label: "全部工具", Click: invoke(func() { a.navigate(-1) })}, mygo.Separator()}
	for i, tool := range tools {
		i, tool := i, tool
		toolMenu = append(toolMenu, &mygo.MenuItem{Label: tool.title, Click: invoke(func() {
			if a.enabled[tool.id] {
				a.navigate(i)
			} else {
				a.navigate(-2)
			}
		})})
	}
	toolMenu = append(toolMenu, mygo.Separator(), &mygo.MenuItem{Label: "插件管理…", Accelerator: "CmdOrCtrl+,", Click: invoke(func() { a.navigate(-2) })})
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "文件", Submenu: []*mygo.MenuItem{{Label: "选择文件…", Accelerator: "CmdOrCtrl+O", Click: invoke(a.openCurrent)}, {Label: "导出或保存…", Accelerator: "CmdOrCtrl+S", Click: invoke(a.exportCurrent)}, mygo.Separator(), {Role: mygo.RoleClose}}},
		{Role: mygo.RoleEditMenu}, {Label: "工具", Submenu: toolMenu}, {Role: mygo.RoleWindowMenu},
	}))
}
