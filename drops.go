package main

import (
	"path/filepath"
	"strings"
)

func (a *app) acceptDrop(paths []string) {
	if len(paths) == 0 || a.working || a.ocrBusy {
		return
	}
	id := ""
	if a.page >= 0 && a.page < len(tools) {
		id = tools[a.page].id
	}
	if id == "" || id == "templates" {
		switch strings.ToLower(filepath.Ext(paths[0])) {
		case ".md", ".markdown":
			id = "markdown"
		default:
			if fileConversionKind(paths[0]) == "图片" {
				id = "ocr"
			} else {
				id = "convert"
			}
		}
	}
	for i, tool := range tools {
		if tool.id == id {
			if !a.enabled[id] {
				a.status = "请先启用对应插件"
				return
			}
			a.navigate(i)
			break
		}
	}
	switch id {
	case "ocr":
		for _, path := range paths {
			if fileConversionKind(path) != "图片" {
				a.status = "请拖入图片"
				return
			}
		}
		a.ocrFiles = paths
		a.ocrSources = append([]string(nil), paths...)
		a.ocrText = ""
		a.selectOCR(0)
	case "photos":
		for _, path := range paths {
			if fileConversionKind(path) != "图片" {
				a.status = "请拖入图片"
				return
			}
		}
		a.photoPreview = nil
		for _, path := range paths {
			a.photos = append(a.photos, photo{Path: path, Original: path})
		}
		if err := a.loadPhoto(len(a.photos) - 1); err != nil {
			a.status = err.Error()
		}
	case "markdown":
		ext := strings.ToLower(filepath.Ext(paths[0]))
		if ext != ".md" && ext != ".markdown" && ext != ".txt" {
			a.status = "请拖入 Markdown 文件"
			return
		}
		a.md = paths[0]
		a.preview = nil
	case "pptx":
		if strings.EqualFold(filepath.Ext(paths[0]), ".pdf") {
			if a.notesKind == "Typst 源文件" {
				a.notesKind = "不使用备注"
			}
			a.setPresentationPDF(paths[0])
		} else if strings.EqualFold(filepath.Ext(paths[0]), ".typ") {
			a.notesKind = "Typst 源文件"
			a.source = paths[0]
		} else {
			a.status = "请拖入 PDF 或 Typst 文件"
		}
	case "convert":
		a.acceptConversionFiles(paths)
	}
}
