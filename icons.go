package main

import "github.com/egoist/mygo/ui"

var toolIcons = map[string]*ui.SVG{
	"folder":  iconSVG(`<path d="M3 7V5a1 1 0 0 1 1-1h5l2 3h9a1 1 0 0 1 1 1v11H3z"/>`),
	"close":   iconSVG(`<path d="m6 6 12 12M18 6 6 18"/>`),
	"up":      iconSVG(`<path d="m6 14 6-6 6 6"/>`),
	"down":    iconSVG(`<path d="m6 10 6 6 6-6"/>`),
	"rotate":  iconSVG(`<path d="M20 7v6h-6M20 12a8 8 0 1 0-2 6"/>`),
	"back":    iconSVG(`<path d="m14 5-7 7 7 7"/>`),
	"refresh": iconSVG(`<path d="M20 4v6h-6M4 20v-6h6M20 10a8 8 0 0 0-14-5M4 14a8 8 0 0 0 14 5"/>`),
	"open":    iconSVG(`<path d="M14 3h7v7m0-7L10 14M9 5H4v15h15v-5"/>`),
	"check":   iconSVG(`<path d="m5 12 4 4L19 6"/>`),

	"home":      iconSVG(`<path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1z"/>`),
	"markdown":  iconSVG(`<path d="M6 3h8l4 4v14H6zM14 3v5h4M9 12h6M9 16h6"/>`),
	"photos":    iconSVG(`<rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8" cy="9" r="1.5"/><path d="m3 17 5-5 4 4 4-6 5 7"/>`),
	"pptx":      iconSVG(`<rect x="3" y="3" width="18" height="13" rx="2"/><path d="M12 16v5m-5 0 5-5 5 5M7 8h10M7 12h6"/>`),
	"templates": iconSVG(`<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 9v12"/>`),
	"convert":   iconSVG(`<path d="M4 7h15m-4-4 4 4-4 4M20 17H5m4-4-4 4 4 4"/>`),
	"ocr":       iconSVG(`<path d="M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5M8 8h8m-4 0v9m-3 0h6"/>`),
	"settings":  iconSVG(`<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="9" cy="6" r="2"/><circle cx="15" cy="12" r="2"/><circle cx="9" cy="18" r="2"/>`),
}

func iconSVG(paths string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`))
}
