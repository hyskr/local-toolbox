package main

import (
	"testing"
)

func TestPreferencesPersist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &app{margin: 18, line: 1.9, spacing: 14, fontsize: 16, width: 145, gap: 45, resolution: 1920, mode: "灰度", template: "课程报告", notesKind: "不使用备注"}
	a.loadPreferences()
	a.margin = 12
	a.template = "科研幻灯片"
	a.conversionKind = "PDF"
	a.conversionFormat = "MD"
	a.persistPreferences()
	b := &app{}
	b.loadPreferences()
	if b.margin != 12 || b.template != "科研幻灯片" || b.formatByKind["PDF"] != "MD" {
		t.Fatalf("preferences not restored: %#v", b.preferenceValue())
	}
}

func TestAvailableFormats(t *testing.T) {
	formats := availableFormats("图片", []string{"one.jpeg", "two.jpg"})
	for _, format := range formats {
		if format == "JPG" {
			t.Fatal("input format should not be offered as output")
		}
	}
	if len(formats) == 0 {
		t.Fatal("expected alternative output formats")
	}
}
