package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *app) loadTemplatePreviews() {
	home, _ := os.UserHomeDir()
	for i, name := range []string{"course-report", "sci-brain-slides-zh"} {
		if a.templatesPreview[i] != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(home, "Documents/typst-templates/templates", name, "thumbnail.png"))
		if err == nil {
			a.templatesPreview[i], _ = ui.DecodeBitmap(data)
		}
	}
}

func (a *app) createTemplate() {
	folder, name, choice := a.folder, strings.TrimSpace(a.project), a.template
	a.job("创建项目", func() (string, error) {
		if folder == "" || name == "" || filepath.Base(name) != name || name == "." || name == ".." {
			return "", fmt.Errorf("请选择文件夹并填写项目名称")
		}
		home, _ := os.UserHomeDir()
		repo := filepath.Join(home, "Documents/typst-templates")
		pkg, version := "course-report", "0.1.0"
		if choice == "科研幻灯片" {
			pkg, version = "sci-brain-slides-zh", "0.1.1"
		}
		if err := run("sh", filepath.Join(repo, "scripts/install.sh"), pkg); err != nil {
			return "", err
		}
		out := filepath.Join(folder, name)
		if err := run("typst", "init", "@local/"+pkg+":"+version, out); err != nil {
			return "", err
		}
		return out, nil
	})
}
func (a *app) templates(c *ui.Context) {
	pageHeader(c, "templates", "Typst 模板", func() {
		if ui.PrimaryButton(c, "创建项目").Disabled(a.folder == "" || strings.TrimSpace(a.project) == "" || a.working).Clicked() {
			a.createTemplate()
		}
	})
	workspace(c, func() {
		section(c, "模板", func() {
			for _, name := range []string{"课程报告", "科研幻灯片"} {
				name := name
				row := ui.ButtonBase(c).FillWidth().Height(42).Padding(0, 12).Gap(10).Justify(ui.Start).Radius(7).Label(name).Disabled(a.working)
				if a.template == name {
					row.Background(colors(c).selected)
				} else if row.Hovered() {
					row.Background(c.Theme().SurfaceHover)
				}
				row.Children(func() {
					ui.Icon(c, toolIcons["templates"]).Size(17, 17).TextColor(c.Theme().Accent)
					ui.Text(c, name).Grow(1).FontSize(13)
					if a.template == name {
						ui.Icon(c, toolIcons["check"]).Size(15, 15).TextColor(c.Theme().Accent)
					}
				})
				row.OnClick(func() { a.template = name })
			}
		})
		ui.Divider(c)
		section(c, "项目名称", func() {
			ui.TextInput(c, &a.project).FillWidth().Placeholder("输入项目名称").Label("项目名称").Disabled(a.working)
		})
		section(c, "保存位置", func() {
			if a.folder != "" {
				compactPath(c, a.folder)
			}
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "选择文件夹…").Disabled(a.working).Clicked() {
					a.pickDir(func(path string) { a.folder = path })
				}
				if a.folder != "" && iconButton(c, "folder", "打开文件夹", false).Clicked() {
					if err := exec.Command("open", a.folder).Start(); err != nil {
						a.status = err.Error()
					}
				}
			})
		})
	}, func() {
		paneBar(c, "模板预览", nil)
		index := 0
		if a.template == "科研幻灯片" {
			index = 1
		}
		if a.templatesPreview[index] != nil {
			paperPreview(c, a.templatesPreview[index])
		} else {
			emptyStage(c, "templates", a.template, "", nil)
		}
	})
}
