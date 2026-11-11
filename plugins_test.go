package main

import (
	"os"
	"testing"
)

func TestPluginSettingsPersist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	initial := loadPlugins()
	for _, plugin := range tools {
		if !initial[plugin.id] {
			t.Fatalf("%s should start enabled", plugin.id)
		}
	}
	initial["photos"] = false
	if err := savePlugins(initial); err != nil {
		t.Fatal(err)
	}
	restored := loadPlugins()
	if restored["photos"] || !restored["markdown"] {
		t.Fatalf("unexpected settings: %v", restored)
	}
	path, err := pluginSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"photos":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	restored = loadPlugins()
	if restored["photos"] || !restored["templates"] {
		t.Fatalf("missing plugin should default to enabled: %v", restored)
	}
}
