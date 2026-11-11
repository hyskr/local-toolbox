package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type conversionItem struct {
	Path    string   `json:"path"`
	State   string   `json:"state"`
	Outputs []string `json:"outputs"`
	Error   string   `json:"error"`
}

func conversionStateLabel(s string) string {
	switch s {
	case "pending":
		return "等待中"
	case "running":
		return "转换中"
	case "done":
		return "完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已停止"
	}
	return s
}
func isTextResult(path string) bool {
	p := strings.ToLower(path)
	return strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".txt")
}
func hasFailedConversions(items []conversionItem) bool {
	for _, v := range items {
		if v.State == "failed" {
			return true
		}
	}
	return false
}
func (a *app) startConversion() {
	a.startConversionPaths(append([]string(nil), a.conversionFiles...), false)
}
func (a *app) startConversionPaths(paths []string, retry bool) {
	if a.working || len(paths) == 0 {
		return
	}
	if !retry {
		a.conversionItems = nil
		a.conversionOutput = ""
		a.conversionText = ""
		for _, path := range paths {
			a.conversionItems = append(a.conversionItems, conversionItem{Path: path, State: "pending"})
		}
	} else {
		for i := range a.conversionItems {
			if a.conversionItems[i].State == "failed" {
				a.conversionItems[i].State = "pending"
				a.conversionItems[i].Error = ""
			}
		}
	}
	a.working = true
	owner := a.page
	a.jobPage = owner
	a.status = ""
	a.last = ""
	ctx, cancel := context.WithCancel(context.Background())
	a.conversionCancel = cancel
	format := strings.ToLower(a.conversionFormat)
	folder := a.conversionFolder
	go func() {
		defer cancel()
		script, err := a.converterScript()
		var diagnostics bytes.Buffer
		if err == nil {
			cmd := exec.CommandContext(ctx, "/usr/bin/swift", append([]string{"-suppress-warnings", script, format}, paths...)...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) }
			cmd.Stderr = &diagnostics
			cmd.Env = append(os.Environ(), "TOOLBOX_OUTPUT_DIRECTORY="+folder)
			pipe, e := cmd.StdoutPipe()
			err = e
			if err == nil {
				err = cmd.Start()
			}
			if err == nil {
				scanner := bufio.NewScanner(pipe)
				scanner.Buffer(make([]byte, 4096), 1024*1024)
				for scanner.Scan() {
					var event conversionItem
					if json.Unmarshal(scanner.Bytes(), &event) != nil {
						continue
					}
					a.win.Update(func() {
						for i := range a.conversionItems {
							if a.conversionItems[i].Path == event.Path {
								a.conversionItems[i] = event
								break
							}
						}
					})
				}
				scanErr := scanner.Err()
				err = cmd.Wait()
				if err == nil {
					err = scanErr
				}
			}
		}
		cancelled := ctx.Err() != nil
		a.win.Update(func() {
			a.working = false
			a.conversionCancel = nil
			for i := range a.conversionItems {
				item := &a.conversionItems[i]
				if item.State == "running" || item.State == "pending" {
					if cancelled {
						item.State = "cancelled"
					} else {
						item.State = "failed"
						item.Error = diagnostics.String()
						if item.Error == "" && err != nil {
							item.Error = err.Error()
						}
					}
				}
			}
			if err != nil && !cancelled {
				a.setNotice(owner, "转换进程未完成，请查看任务错误", "")
			}
		})
	}()
}
