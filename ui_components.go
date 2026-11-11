package main

import (
	"github.com/egoist/mygo/ui"
	"path/filepath"
)

type uiColors struct{ surface, sidebar, canvas, border, selected ui.Color }

func colors(c *ui.Context) uiColors {
	if c.Theme().Dark {
		return uiColors{ui.Hex("#22252b"), ui.Hex("#1b1e23"), ui.Hex("#171a1f"), ui.Hex("#343943"), ui.Hex("#2c3d58")}
	}
	return uiColors{ui.Hex("#ffffff"), ui.Hex("#f2f5f9"), ui.Hex("#edf1f6"), ui.Hex("#dfe5ee"), ui.Hex("#e4edfb")}
}
func iconButton(c *ui.Context, id, label string, disabled bool) ui.Element {
	p := colors(c)
	b := ui.ButtonBase(c).Size(30, 30).Radius(6).Center().Label(label).Tooltip(label).Disabled(disabled)
	if b.Hovered() {
		b.Background(p.selected)
	}
	return b.Children(func() { ui.Icon(c, toolIcons[id]).Size(16, 16).TextColor(c.Theme().TextMuted) })
}
func pageHeader(c *ui.Context, id, title string, actions func()) {
	ui.Row(c).FillWidth().Height(68).Shrink(0).Padding(0, 24).Gap(12).Background(colors(c).surface).Children(func() {
		ui.Icon(c, toolIcons[id]).Size(21, 21).TextColor(c.Theme().Accent)
		ui.Text(c, title).FontSize(18).Bold().SingleLine()
		ui.Spacer(c)
		if actions != nil {
			actions()
		}
	})
	ui.Divider(c)
}
func workspace(c *ui.Context, inspector, content func()) {
	ui.Row(c).Grow(1).MinHeight(0).FillWidth().AlignItems(ui.Stretch).Children(func() {
		if inspector != nil {
			ui.Scroll(c).Width(284).Shrink(0).FillHeight().Background(colors(c).surface).Children(func() { ui.Column(c).Padding(20).Gap(18).FillWidth().Children(inspector) })
			ui.Box(c).Width(1).FillHeight().Background(colors(c).border)
		}
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Background(colors(c).canvas).Children(content)
	})
}
func paneBar(c *ui.Context, title string, actions func()) {
	ui.Row(c).FillWidth().Height(48).Shrink(0).Padding(0, 18).Gap(8).Children(func() {
		ui.Text(c, title).FontSize(12).Bold().TextColor(c.Theme().TextMuted)
		ui.Spacer(c)
		if actions != nil {
			actions()
		}
	})
}
func section(c *ui.Context, title string, body func()) {
	ui.Column(c).FillWidth().Gap(10).Children(func() { ui.Text(c, title).FontSize(12).Bold().TextColor(c.Theme().TextMuted); body() })
}
func numberSetting(c *ui.Context, label string, value *float64, lo, hi, step float64) {
	ui.Row(c).FillWidth().Gap(10).Children(func() {
		ui.Text(c, label).Grow(1).FontSize(13)
		ui.NumberInput(c, value, lo, hi, step).Width(110).Label(label)
	})
}
func emptyStage(c *ui.Context, id, title, button string, action func()) {
	ui.Column(c).Grow(1).MinHeight(0).FillWidth().Center().Padding(24).Gap(18).Children(func() {
		ui.Box(c).Size(72, 72).Radius(20).Background(colors(c).surface).Center().Children(func() { ui.Icon(c, toolIcons[id]).Size(32, 32).TextColor(c.Theme().Accent) })
		ui.Text(c, title).FontSize(17).Bold()
		if action != nil && button != "" {
			if ui.Button(c, button).Height(34).Clicked() {
				action()
			}
		}
	})
}
func (a *app) sourceFile(c *ui.Context, path, label string, choose func()) {
	section(c, label, func() {
		if path == "" {
			if ui.Button(c, "选择文件…").FillWidth().Disabled(a.working).Clicked() {
				choose()
			}
			return
		}
		ui.Row(c).FillWidth().Gap(6).Children(func() {
			ui.Text(c, filepath.Base(path)).Grow(1).MinWidth(0).FontSize(13).SingleLine().Tooltip(path)
			if iconButton(c, "folder", "打开源文件夹", false).Clicked() {
				a.openSourceFolder(path)
			}
		})
		if ui.Button(c, "更换文件…").Disabled(a.working).Clicked() {
			choose()
		}
	})
}
func paperPreview(c *ui.Context, bitmap *ui.Bitmap) {
	w, h := bitmap.Size()
	ratio := float32(w) / float32(h)
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
		ui.Column(c).Padding(24).FillWidth().AlignItems(ui.Center).Children(func() { ui.Image(c, bitmap).FillWidth().MaxWidth(650).AspectRatio(ratio).Fit(ui.Contain) })
	})
}
func compactPath(c *ui.Context, path string) {
	ui.Text(c, path).FontSize(12).TextColor(c.Theme().TextMuted).SingleLine().Tooltip(path)
}
