package main

import (
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
)

func (a *app) photoRun(cfg any) error {
	b, e := json.Marshal(cfg)
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(a.temp, "photo-*.json")
	if e != nil {
		return e
	}
	p := f.Name()
	defer os.Remove(p)
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	f.Close()
	script, e := a.script("photo.py")
	if e != nil {
		return e
	}
	return run("python3", script, p)
}
func (a *app) loadPhoto(i int) error {
	b, e := os.ReadFile(a.photos[i].Path)
	if e != nil {
		return e
	}
	bm, e := ui.DecodeBitmap(b)
	if e != nil {
		return e
	}
	w, h := bm.Size()
	a.bitmap = bm
	a.imageW = float64(w)
	a.imageH = float64(h)
	a.selected = i
	a.points = nil
	return nil
}
func (a *app) pickPhotos() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Title: "添加照片", Multiple: true, Filters: []mygo.FileFilter{{Name: "照片", Extensions: []string{"png", "jpg", "jpeg", "webp", "bmp"}}}})
		a.win.Update(func() {
			if err != nil {
				a.status = err.Error()
				return
			}
			if a.working {
				return
			}
			if len(paths) > 0 {
				a.photoPreview = nil
				for _, path := range paths {
					a.photos = append(a.photos, photo{Path: path, Original: path})
				}
				if err = a.loadPhoto(len(a.photos) - 1); err != nil {
					a.status = err.Error()
				}
			}
		})
	}()
}
func (a *app) removePhoto() {
	if a.working || len(a.photos) == 0 {
		return
	}
	i := a.selected
	a.photos = append(a.photos[:i], a.photos[i+1:]...)
	a.photoPreview = nil
	a.bitmap = nil
	a.points = nil
	if len(a.photos) > 0 {
		if i >= len(a.photos) {
			i = len(a.photos) - 1
		}
		if err := a.loadPhoto(i); err != nil {
			a.status = err.Error()
		}
	}
}
func (a *app) movePhoto(delta int) {
	i, j := a.selected, a.selected+delta
	if a.working || j < 0 || j >= len(a.photos) {
		return
	}
	a.photos[i], a.photos[j] = a.photos[j], a.photos[i]
	a.photoPreview = nil
	if err := a.loadPhoto(j); err != nil {
		a.status = err.Error()
	}
}
func (a *app) transformPhoto(kind string) {
	if a.working || a.selected >= len(a.photos) {
		return
	}
	if kind == "scan" && len(a.points) != 0 && len(a.points) != 4 {
		a.status = "请选择四个角，或清除选点后处理整张图片"
		return
	}
	i := a.selected
	input := a.photos[i].Path
	points := append([][2]float64{}, a.points...)
	mode, shadow := a.mode, a.shadow
	label := "旋转照片"
	if kind == "scan" {
		label = "扫描照片"
	}
	a.job(label, func() (string, error) {
		file, err := os.CreateTemp(a.temp, kind+"-*.jpg")
		if err != nil {
			return "", err
		}
		out := file.Name()
		file.Close()
		cfg := map[string]any{"action": kind, "input": input, "output": out, "points": points, "mode": mode, "shadow": shadow}
		if err = a.photoRun(cfg); err != nil {
			return "", err
		}
		a.win.Update(func() {
			a.photos[i].Path = out
			a.photos[i].Rotation = 0
			a.photoPreview = nil
			if err := a.loadPhoto(i); err != nil {
				a.status = err.Error()
			}
		})
		return out, nil
	})
}
func (a *app) previewPhotos() {
	items := append([]photo(nil), a.photos...)
	width, gap, one := a.width, a.gap, a.onePage
	a.job("生成排版预览", func() (string, error) {
		out := filepath.Join(a.temp, "photo-preview.pdf")
		if err := a.photoRun(map[string]any{"action": "export", "photos": items, "width": width, "gap": gap, "onePage": one, "output": out}); err != nil {
			return "", err
		}
		bitmap, err := a.pdfImage(out)
		if err != nil {
			return "", err
		}
		a.win.Update(func() { a.photoPreview = bitmap })
		return out, nil
	})
}
func (a *app) exportPhotos() {
	items := append([]photo(nil), a.photos...)
	width, gap, one := a.width, a.gap, a.onePage
	a.save("导出照片 PDF", "照片笔记-A4.pdf", "pdf", func(out string) (string, error) {
		return out, a.photoRun(map[string]any{"action": "export", "photos": items, "width": width, "gap": gap, "onePage": one, "output": out})
	})
}
func (a *app) photoPage(c *ui.Context) {
	pageHeader(c, "photos", "照片 A4 与扫描", func() {
		if len(a.photos) > 0 {
			if ui.Button(c, "添加照片…").Disabled(a.working).Clicked() {
				a.pickPhotos()
			}
			if ui.PrimaryButton(c, "导出 PDF…").Disabled(a.working).Clicked() {
				a.exportPhotos()
			}
		}
	})
	if len(a.photos) == 0 {
		emptyStage(c, "photos", "拖入照片，排成 A4", "添加照片…", a.pickPhotos)
		return
	}
	workspace(c, func() {
		section(c, fmt.Sprintf("照片 · %d 张", len(a.photos)), func() {
			height := float32(len(a.photos) * 40)
			if height > 205 {
				height = 205
			}
			ui.Scroll(c).FillWidth().Height(height).Children(func() {
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					for i, p := range a.photos {
						i, p := i, p
						row := ui.ButtonBase(c).FillWidth().Height(36).Padding(0, 9).Gap(8).Radius(6).Justify(ui.Start).Disabled(a.working).Label(filepath.Base(p.Original))
						if i == a.selected {
							row.Background(colors(c).selected)
						} else if row.Hovered() {
							row.Background(c.Theme().SurfaceHover)
						}
						row.Children(func() {
							ui.Text(c, fmt.Sprint(i+1)).Width(16).FontSize(11).TextColor(c.Theme().TextMuted)
							ui.Text(c, filepath.Base(p.Original)).Grow(1).MinWidth(0).SingleLine().FontSize(12).Tooltip(p.Original)
						})
						row.OnClick(func() {
							if err := a.loadPhoto(i); err != nil {
								a.status = err.Error()
							}
						})
					}
				})
			})
			ui.Row(c).Gap(4).Children(func() {
				if iconButton(c, "up", "上移", a.working || a.selected == 0).Clicked() {
					a.movePhoto(-1)
				}
				if iconButton(c, "down", "下移", a.working || a.selected == len(a.photos)-1).Clicked() {
					a.movePhoto(1)
				}
				ui.Spacer(c)
				if iconButton(c, "folder", "打开源文件夹", false).Clicked() {
					a.openSourceFolder(a.photos[a.selected].Original)
				}
				if iconButton(c, "close", "移除照片", a.working).Clicked() {
					a.removePhoto()
				}
			})
		})
		ui.Divider(c)
		if a.photoView == 0 {
			oldW, oldG, oldO := a.width, a.gap, a.onePage
			section(c, "A4 排版", func() {
				numberSetting(c, "图片宽度 · mm", &a.width, 80, 186, 1)
				numberSetting(c, "图片后留白 · mm", &a.gap, 15, 180, 1)
				ui.Checkbox(c, &a.onePage, "每张照片单独一页")
			})
			if oldW != a.width || oldG != a.gap || oldO != a.onePage {
				a.photoPreview = nil
			}
		} else {
			section(c, "扫描效果", func() {
				ui.Select(c, &a.mode, []string{"原色", "灰度", "黑白"}).FillWidth().Disabled(a.working)
				ui.Checkbox(c, &a.shadow, "去阴影").Disabled(a.working)
			})
			if ui.Button(c, "应用扫描").FillWidth().Disabled(a.working || (len(a.points) != 0 && len(a.points) != 4)).Clicked() {
				a.transformPhoto("scan")
			}
			if ui.Button(c, "恢复原图").FillWidth().Disabled(a.working).Clicked() {
				a.photos[a.selected].Path = a.photos[a.selected].Original
				a.photoPreview = nil
				if err := a.loadPhoto(a.selected); err != nil {
					a.status = err.Error()
				}
			}
		}
	}, func() {
		ui.Row(c).FillWidth().Height(52).Shrink(0).Padding(0, 16).Gap(8).Children(func() {
			ui.Segmented(c, &a.photoView, "排版预览", "扫描校正").Label("照片视图")
			ui.Spacer(c)
			if a.photoView == 0 {
				if ui.Button(c, "更新预览").Disabled(a.working).Clicked() {
					a.previewPhotos()
				}
			} else {
				if iconButton(c, "rotate", "旋转 90°", a.working).Clicked() {
					a.transformPhoto("rotate")
				}
				if ui.Button(c, "识别文字").Disabled(a.working).Clicked() {
					a.openOCR([]string{a.photos[a.selected].Path})
				}
			}
		})
		if a.photoView == 0 {
			if a.photoPreview == nil {
				emptyStage(c, "photos", "预览 A4 排版", "生成预览", func() {
					if !a.working {
						a.previewPhotos()
					}
				})
			} else {
				paperPreview(c, a.photoPreview)
			}
			return
		}
		a.scanCanvas(c)
	})
}
func (a *app) scanCanvas(c *ui.Context) {
	if a.bitmap == nil {
		emptyStage(c, "photos", "无法预览这张照片", "", nil)
		return
	}
	ui.Row(c).FillWidth().Padding(0, 18, 10, 18).Gap(8).Children(func() {
		ui.Text(c, fmt.Sprintf("按左上、右上、右下、左下选点 · %d/4", len(a.points))).FontSize(12).TextColor(c.Theme().TextMuted)
		ui.Spacer(c)
		if ui.Button(c, "清除选点").Disabled(a.working).Clicked() {
			a.points = nil
		}
	})
	viewW, viewH := c.Size()
	maxW, maxH := viewW-535, viewH-230
	if maxW < 280 {
		maxW = 280
	}
	if maxH < 240 {
		maxH = 240
	}
	w := maxW
	h := float32(a.imageH/a.imageW) * w
	if h > maxH {
		w = w * maxH / h
		h = maxH
	}
	ui.Column(c).Grow(1).MinHeight(0).FillWidth().Center().Children(func() {
		el := ui.Image(c, a.bitmap).Size(w, h).Fit(ui.Contain).OnClick(func() {}).Label("选择文稿四角").Disabled(a.working)
		x, y, _ := el.PointerPosition()
		if el.Clicked() && len(a.points) < 4 {
			a.points = append(a.points, [2]float64{float64(x / w), float64(y / h)})
		}
		points := append([][2]float64(nil), a.points...)
		accent := c.Theme().Accent
		el.DrawOver(func(p *ui.Painter, r ui.Rect) {
			var outline ui.Path
			for i, pt := range points {
				x, y := r.X+float32(pt[0])*r.W, r.Y+float32(pt[1])*r.H
				if i == 0 {
					outline.MoveTo(x, y)
				} else {
					outline.LineTo(x, y)
				}
				p.Fill(ui.Rect{X: x - 5, Y: y - 5, W: 10, H: 10}, accent, 5)
			}
			if len(points) == 4 {
				outline.LineTo(r.X+float32(points[0][0])*r.W, r.Y+float32(points[0][1])*r.H)
			}
			if len(points) > 1 {
				p.StrokePath(&outline, 2, accent)
			}
		})
	})
}
