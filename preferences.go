package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type preferences struct {
	TemplateFolder                              string
	Compact                                     bool `json:"compact"`
	Margin, Line, Spacing, FontSize             float64
	Width, Gap, Resolution                      float64
	OnePage, Shadow                             bool
	Mode, Template, NotesKind, ConversionFolder string
	Formats                                     map[string]string
}

func preferencesPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "local-toolbox", "preferences.json"), nil
}
func (a *app) preferenceValue() preferences {
	return preferences{TemplateFolder: a.folder, Compact: a.compact, Margin: a.margin, Line: a.line, Spacing: a.spacing, FontSize: a.fontsize, Width: a.width, Gap: a.gap, Resolution: a.resolution, OnePage: a.onePage, Shadow: a.shadow, Mode: a.mode, Template: a.template, NotesKind: a.notesKind, ConversionFolder: a.conversionFolder, Formats: a.formatByKind}
}
func (a *app) loadPreferences() {
	a.formatByKind = map[string]string{}
	path, err := preferencesPath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	p := a.preferenceValue()
	if json.Unmarshal(data, &p) != nil {
		return
	}
	a.compact = p.Compact
	a.margin = p.Margin
	a.line = p.Line
	a.spacing = p.Spacing
	a.fontsize = p.FontSize
	a.width = p.Width
	a.gap = p.Gap
	a.resolution = p.Resolution
	a.onePage = p.OnePage
	a.shadow = p.Shadow
	a.mode = p.Mode
	a.template = p.Template
	a.notesKind = p.NotesKind
	a.conversionFolder = p.ConversionFolder
	a.folder = p.TemplateFolder
	if p.Formats != nil {
		a.formatByKind = p.Formats
	}
}
func (a *app) persistPreferences() {
	if a.conversionKind != "" && a.conversionFormat != "" {
		a.formatByKind[a.conversionKind] = a.conversionFormat
	}
	data, err := json.Marshal(a.preferenceValue())
	if err != nil || string(data) == a.preferencesSnapshot {
		return
	}
	path, err := preferencesPath()
	if err == nil {
		err = os.MkdirAll(filepath.Dir(path), 0700)
	}
	if err == nil {
		err = os.WriteFile(path, data, 0600)
	}
	if err != nil {
		a.status = "无法保存偏好：" + err.Error()
		return
	}
	a.preferencesSnapshot = string(data)
}
