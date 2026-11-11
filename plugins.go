package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
)

func pluginSettingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "local-toolbox", "plugins.json"), nil
}
func loadPlugins() map[string]bool {
	enabled := make(map[string]bool, len(tools))
	for _, tool := range tools {
		enabled[tool.id] = true
	}
	path, err := pluginSettingsPath()
	if err != nil {
		return enabled
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return enabled
	}
	var saved map[string]bool
	if json.Unmarshal(data, &saved) == nil {
		for _, tool := range tools {
			if value, ok := saved[tool.id]; ok {
				enabled[tool.id] = value
			}
		}
	}
	return enabled
}
func savePlugins(enabled map[string]bool) error {
	path, err := pluginSettingsPath()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(enabled, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
func (a *app) navItem(c *ui.Context, page int, label, unused string) {
	selected := a.page == page
	row := ui.ButtonBase(c).FillWidth().Height(37).Padding(0, 10).Gap(9).Justify(ui.Start).Radius(7).Label(label).Cursor(ui.CursorPointer)
	if selected {
		row.Background(colors(c).selected)
	} else if row.Hovered() {
		row.Background(c.Theme().SurfaceHover)
	}
	id := "home"
	if page == -2 {
		id = "settings"
	} else if page >= 0 {
		id = tools[page].id
	}
	row.Children(func() {
		col := c.Theme().TextMuted
		if selected {
			col = c.Theme().Accent
		}
		ui.Icon(c, toolIcons[id]).Size(17, 17).TextColor(col)
		ui.Text(c, label).Grow(1).MinWidth(0).FontSize(12).SingleLine()
	})
	row.OnClick(func() { a.navigate(page) })
}
func (a *app) pluginsPage(c *ui.Context) {
	pageHeader(c, "settings", "插件管理", nil)
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Background(colors(c).canvas).Children(func() {
		ui.Column(c).Padding(28).Gap(10).FillWidth().Children(func() {
			for _, tool := range tools {
				tool := tool
				ui.Row(c).FillWidth().MaxWidth(800).Padding(16, 18).Gap(14).Radius(9).Background(colors(c).surface).Children(func() {
					ui.Box(c).Size(38, 38).Radius(10).Background(colors(c).selected).Center().Children(func() { ui.Icon(c, toolIcons[tool.id]).Size(21, 21).TextColor(c.Theme().Accent) })
					ui.Text(c, tool.title).Grow(1).FontSize(14).Bold()
					enabled := a.enabled[tool.id]
					if ui.Switch(c, &enabled).Label("启用 " + tool.title).Changed() {
						a.enabled[tool.id] = enabled
						if err := savePlugins(a.enabled); err != nil {
							a.enabled[tool.id] = !enabled
							a.status = "无法保存设置：" + err.Error()
						} else {
							if tool.id == "templates" && enabled {
								a.loadTemplatePreviews()
							}
							a.status = "设置已保存"
						}
					}
				})
			}
		})
	})
}
