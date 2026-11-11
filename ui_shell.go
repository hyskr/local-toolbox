package main

import (
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"os/exec"
)

type toolNotice struct{ text, path string }

func (a *app) setNotice(page int, text, path string) {
	a.notices[page] = toolNotice{text, path}
	if a.page == page {
		a.status, a.last = text, path
	}
}
func (a *app) navigate(page int) {
	a.notices[a.page] = toolNotice{a.status, a.last}
	a.page = page
	n := a.notices[page]
	a.status, a.last = n.text, n.path
}
func (a *app) view(c *ui.Context) {
	p := colors(c)
	ui.Row(c).Fill().Background(p.surface).Children(func() {
		ui.Column(c).Width(186).Shrink(0).FillHeight().Padding(16, 10).Gap(5).Background(p.sidebar).Children(func() {
			ui.Row(c).Padding(8, 10, 24, 10).Gap(10).Children(func() {
				if a.icon != nil {
					ui.Image(c, a.icon).Size(36, 36)
				}
				ui.Text(c, "工具箱").FontSize(17).Bold()
			})
			a.navItem(c, -1, "全部工具", "")
			ui.Text(c, "转换与导出").FontSize(11).TextColor(c.Theme().TextMuted).Padding(18, 10, 6, 10)
			for _, i := range []int{4, 0, 2} {
				if a.enabled[tools[i].id] {
					a.navItem(c, i, tools[i].title, "")
				}
			}
			ui.Text(c, "图片与写作").FontSize(11).TextColor(c.Theme().TextMuted).Padding(18, 10, 6, 10)
			for _, i := range []int{5, 1, 3} {
				if a.enabled[tools[i].id] {
					a.navItem(c, i, tools[i].title, "")
				}
			}
			ui.Spacer(c)
			a.navItem(c, -2, "插件管理", "")
		})
		ui.Box(c).Width(1).FillHeight().Background(p.border)
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Children(func() {
			ui.Column(c).Key(a.page).Grow(1).MinHeight(0).FillWidth().Children(func() {
				if a.page == -1 {
					a.home(c)
				} else if a.page == -2 {
					a.pluginsPage(c)
				} else if a.page >= 0 && a.page < len(tools) && a.enabled[tools[a.page].id] {
					tools[a.page].view(a, c)
				}
			})
			if a.status != "" || a.working || a.ocrBusy {
				a.noticeBar(c)
			}
		})
	})
	a.persistPreferences()
}
func (a *app) noticeBar(c *ui.Context) {
	ui.Divider(c)
	ui.Row(c).FillWidth().MinHeight(42).Padding(8, 20).Gap(10).Children(func() {
		if a.working || a.ocrBusy {
			ui.Spinner(c).Size(15, 15).Label("处理中")
		}
		message := a.status
		if a.working && a.jobPage != a.page && a.jobPage >= 0 && a.jobPage < len(tools) {
			message = tools[a.jobPage].title + " · 处理中…"
		} else if a.ocrBusy && a.page != 5 {
			message = "图片文字识别 · 处理中…"
		}
		if message == "" {
			message = "正在处理…"
		}
		ui.Text(c, message).Grow(1).MinWidth(0).FontSize(12).SingleLine().Tooltip(message)
		if a.working && a.jobPage != a.page {
			if ui.Button(c, "查看任务").Clicked() {
				a.navigate(a.jobPage)
			}
		}
		if a.last != "" {
			if ui.Button(c, "打开结果").Clicked() {
				_ = exec.Command("open", a.last).Start()
			}
			if iconButton(c, "folder", "在 Finder 中显示结果", false).Clicked() {
				_ = exec.Command("open", "-R", a.last).Start()
			}
		}
		if !a.working && !a.ocrBusy && iconButton(c, "close", "关闭提示", false).Clicked() {
			a.status = ""
			a.last = ""
		}
	})
}
func (a *app) home(c *ui.Context) {
	pageHeader(c, "home", "全部工具", func() {
		if ui.Button(c, "打开文件…").Clicked() {
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
		}
	})
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Background(colors(c).canvas).Children(func() {
		ui.Column(c).FillWidth().Padding(28).Gap(22).Children(func() {
			for _, group := range []struct {
				name    string
				indices []int
			}{{"转换与导出", []int{4, 0, 2}}, {"图片与写作", []int{5, 1, 3}}} {
				active := []int{}
				for _, i := range group.indices {
					if a.enabled[tools[i].id] {
						active = append(active, i)
					}
				}
				if len(active) == 0 {
					continue
				}
				ui.Row(c).FillWidth().Children(func() {
					ui.Text(c, group.name).FontSize(13).Bold()
					ui.Spacer(c)
					ui.Text(c, fmt.Sprintf("%d 个工具", len(active))).FontSize(11).TextColor(c.Theme().TextMuted)
				})
				for n := 0; n < len(active); n += 2 {
					n := n
					ui.Row(c).FillWidth().Gap(14).Children(func() {
						a.homeTile(c, active[n])
						if n+1 < len(active) {
							a.homeTile(c, active[n+1])
						} else {
							ui.Box(c).Grow(1).Basis(0)
						}
					})
				}
			}
			count := 0
			for _, v := range a.enabled {
				if v {
					count++
				}
			}
			if count == 0 {
				ui.Text(c, "在插件管理中选择需要的工具。").TextColor(c.Theme().TextMuted)
				if ui.Button(c, "管理插件").Clicked() {
					a.navigate(-2)
				}
			}
		})
	})
}
func (a *app) homeTile(c *ui.Context, i int) {
	tool := tools[i]
	p := colors(c)
	card := ui.ButtonBase(c).Grow(1).Basis(0).Height(92).Padding(20).Gap(14).Justify(ui.Start).Radius(12).Background(p.surface).Border(1, p.border).Label(tool.title).Cursor(ui.CursorPointer)
	if card.Hovered() {
		card.Border(1, c.Theme().Accent)
	}
	card.Children(func() {
		ui.Box(c).Size(42, 42).Radius(12).Background(p.selected).Center().Children(func() { ui.Icon(c, toolIcons[tool.id]).Size(23, 23).TextColor(c.Theme().Accent) })
		ui.Text(c, tool.title).Grow(1).FontSize(15).Bold().SingleLine()
		ui.Text(c, "›").FontSize(22).TextColor(c.Theme().TextMuted)
	})
	card.OnClick(func() { a.navigate(i) })
}
