package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed scripts/*.py scripts/*.swift scripts/fileflipper/*.swift
var scripts embed.FS

//go:embed resources/icon.png
var iconData []byte

type photo struct {
	Path     string `json:"path"`
	Original string `json:"-"`
	Rotation int    `json:"rotation"`
}
type app struct {
	notices        map[int]toolNotice
	jobPage        int
	photoView      int
	ocrSelection   int
	ocrBitmap      *ui.Bitmap
	pptPreview     *ui.Bitmap
	pptPreviewPath string
	pptPreviewBusy bool

	mdPreviewKey                                            string
	formatByKind                                            map[string]string
	preferencesSnapshot                                     string
	conversionFolder                                        string
	conversionItems                                         []conversionItem
	conversionCancel                                        func()
	conversionText, conversionOutput                        string
	working                                                 bool
	ocrFiles                                                []string
	ocrSources                                              []string
	ocrText                                                 string
	ocrBusy                                                 bool
	win                                                     *mygo.Window
	page                                                    int
	md, pdf, source, notes, folder, project, status, last   string
	compact, shadow, onePage                                bool
	margin, line, spacing, fontsize, width, gap, resolution float64
	mode, template, notesKind                               string
	photos                                                  []photo
	selected                                                int
	points                                                  [][2]float64
	bitmap                                                  *ui.Bitmap
	imageW, imageH                                          float64
	preview                                                 *ui.Bitmap
	templatesPreview                                        [2]*ui.Bitmap
	photoPreview                                            *ui.Bitmap
	temp                                                    string
	enabled                                                 map[string]bool
	conversionFiles                                         []string
	conversionKind, conversionFormat                        string
	icon                                                    *ui.Bitmap
}

func runOutput(name string, args ...string) (string, error) {
	home, _ := os.UserHomeDir()
	dirs := []string{filepath.Join(home, ".local/share/mise/shims"), "/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/bin"}
	executable := name
	if !strings.ContainsRune(name, os.PathSeparator) {
		for _, dir := range dirs {
			path := filepath.Join(dir, name)
			if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
				executable = path
				break
			}
		}
	}
	cmd := exec.Command(executable, args...)
	cmd.Env = append(os.Environ(), "PATH="+strings.Join(append(dirs, os.Getenv("PATH")), ":"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s：%s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
func run(name string, args ...string) error { _, err := runOutput(name, args...); return err }

func (a *app) script(name string) (string, error) {
	b, e := scripts.ReadFile("scripts/" + name)
	if e != nil {
		return "", e
	}
	p := filepath.Join(a.temp, name)
	return p, os.WriteFile(p, b, 0600)
}
func (a *app) job(label string, work func() (string, error)) {
	if a.working {
		return
	}
	a.working = true
	owner := a.page
	a.jobPage = owner
	a.last = ""
	a.status = label + "…"
	a.win.Invalidate()
	go func() {
		result, e := work()
		a.win.Update(func() {
			a.working = false
			if e != nil {
				a.setNotice(owner, "操作失败："+e.Error(), "")
			} else {
				a.setNotice(owner, "已完成 · "+filepath.Base(result), result)
			}
		})
	}()
}
func (a *app) pickFile(title string, exts []string, set func(string)) {
	go func() {
		paths, e := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: title, Filters: []mygo.FileFilter{{Name: title, Extensions: exts}}})
		if e == nil && len(paths) > 0 {
			a.win.Update(func() { set(paths[0]) })
		}
	}()
}
func (a *app) pickDir(set func(string)) {
	go func() {
		p, e := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: "选择文件夹", Directory: true})
		if e == nil && len(p) > 0 {
			a.win.Update(func() { set(p[0]) })
		}
	}()
}
func (a *app) save(title, def, ext string, work func(string) (string, error)) {
	go func() {
		p, e := mygo.Dialog.Save(mygo.SaveDialogOptions{Parent: a.win, Title: title, DefaultPath: def, Filters: []mygo.FileFilter{{Name: ext, Extensions: []string{ext}}}})
		if e != nil {
			a.win.Update(func() { a.status = e.Error() })
			return
		}
		if p != "" {
			a.win.Update(func() { a.job(title, func() (string, error) { return work(p) }) })
		}
	}()
}
func (a *app) renderPreview(path string) error {
	prefix := filepath.Join(a.temp, "preview")
	if err := run("pdftoppm", "-f", "1", "-l", "1", "-singlefile", "-scale-to", "1000", "-png", path, prefix); err != nil {
		return err
	}
	data, e := os.ReadFile(prefix + ".png")
	if e != nil {
		return e
	}
	bitmap, e := ui.DecodeBitmap(data)
	if e != nil {
		return e
	}
	a.win.Update(func() { a.preview = bitmap })
	return nil
}
func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

type toolInfo struct {
	id, title, symbol, group string
	color                    ui.Color
	view                     func(*app, *ui.Context)
}

var tools = []toolInfo{
	{"markdown", "Markdown 转 PDF", "文", "文稿", ui.Hex("#56679a"), nil},
	{"photos", "照片 A4 与扫描", "图", "图片", ui.Hex("#468679"), nil},
	{"pptx", "PDF 转 PPTX", "页", "幻灯片", ui.Hex("#ab715c"), nil},
	{"templates", "Typst 模板", "版", "模板", ui.Hex("#8d70a3"), nil},
	{"convert", "格式转换", "换", "转换", ui.Hex("#bd765b"), nil},
	{"ocr", "图片文字识别", "识", "文字", ui.Hex("#377e8a"), nil},
}

func init() {
	tools[0].view = (*app).markdown
	tools[1].view = (*app).photoPage
	tools[2].view = (*app).pptx
	tools[3].view = (*app).templates
	tools[4].view = (*app).convertPage
	tools[5].view = (*app).ocrPage
}

func (a *app) openSourceFolder(path string) {
	if path == "" {
		return
	}
	if err := exec.Command("open", filepath.Dir(path)).Start(); err != nil {
		a.status = "无法打开文件夹：" + err.Error()
	}
}
func main() {
	temp, e := os.MkdirTemp("", "toolbox-")
	if e != nil {
		log.Fatal(e)
	}
	defer os.RemoveAll(temp)
	a := &app{page: -1, temp: temp, margin: 18, line: 1.9, spacing: 14, fontsize: 16, width: 145, gap: 45, resolution: 1920, mode: "灰度", shadow: true, template: "课程报告", notesKind: "不使用备注"}
	a.loadPreferences()
	if data, err := json.Marshal(a.preferenceValue()); err == nil {
		a.preferencesSnapshot = string(data)
	}
	a.enabled = loadPlugins()
	a.notices = map[int]toolNotice{}
	mygo.App.WhenReady(func() {
		a.icon, _ = ui.DecodeBitmap(iconData)
		if a.enabled["templates"] {
			a.loadTemplatePreviews()
		}
		a.win = mygo.NewWindow(mygo.WindowOptions{Title: "工具箱", Width: 1240, Height: 820, MinWidth: 1000, MinHeight: 640, StateKey: "toolbox-main", Content: ui.View(a.view)})
		a.installMenu()
		a.win.OnFileDrop(func(e *mygo.FileDropEvent) {
			paths := append([]string(nil), e.Paths...)
			a.win.Update(func() { a.acceptDrop(paths) })
		})
		a.win.OnClosed(func() { mygo.App.Quit() })
	})
	if e := mygo.App.Run(); e != nil {
		log.Fatal(e)
	}
}
