package main

import (
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

// Theme 是所有自绘组件的唯一配色来源。组件自己不挑颜色，切换主题只需
// 换掉 Scene.theme，所有区域（画布 / 工具栏 / 面板 / 菜单 / 弹窗）同步生效。
type Theme struct {
	Dark       bool
	Bg         types.TColor // 窗口底色 / 画布留白
	Surface    types.TColor // 卡片、面板
	SurfaceAlt types.TColor // 次级块、hover
	Border     types.TColor
	Grid       types.TColor
	GridMajor  types.TColor
	Text       types.TColor
	Muted      types.TColor
	Accent     types.TColor
	AccentSoft types.TColor
	AccentDown types.TColor
	Sel        types.TColor
	SelBorder  types.TColor
	Danger     types.TColor
	Warn       types.TColor
	Success    types.TColor
	PK         types.TColor
	FK         types.TColor
	Line       types.TColor
}

func rgb(r, g, b byte) types.TColor { return colors.RGBToColor(r, g, b) }

func lightTheme() Theme {
	return Theme{
		Dark:       false,
		Bg:         rgb(243, 246, 251),
		Surface:    rgb(255, 255, 255),
		SurfaceAlt: rgb(234, 240, 249),
		Border:     rgb(205, 215, 230),
		Grid:       rgb(226, 233, 243),
		GridMajor:  rgb(208, 219, 234),
		Text:       rgb(30, 38, 52),
		Muted:      rgb(110, 124, 145),
		Accent:     rgb(47, 111, 224),
		AccentSoft: rgb(223, 235, 255),
		AccentDown: rgb(30, 82, 178),
		Sel:        rgb(255, 244, 214),
		SelBorder:  rgb(232, 165, 48),
		Danger:     rgb(206, 61, 61),
		Warn:       rgb(214, 141, 40),
		Success:    rgb(38, 150, 100),
		PK:         rgb(205, 143, 40),
		FK:         rgb(122, 84, 204),
		Line:       rgb(92, 126, 188),
	}
}

func darkTheme() Theme {
	return Theme{
		Dark:       true,
		Bg:         rgb(24, 28, 36),
		Surface:    rgb(36, 42, 53),
		SurfaceAlt: rgb(47, 55, 69),
		Border:     rgb(68, 79, 97),
		Grid:       rgb(38, 45, 57),
		GridMajor:  rgb(50, 60, 76),
		Text:       rgb(231, 237, 245),
		Muted:      rgb(142, 155, 174),
		Accent:     rgb(86, 148, 255),
		AccentSoft: rgb(39, 62, 100),
		AccentDown: rgb(56, 108, 214),
		Sel:        rgb(96, 76, 32),
		SelBorder:  rgb(232, 176, 74),
		Danger:     rgb(226, 92, 92),
		Warn:       rgb(230, 166, 72),
		Success:    rgb(72, 190, 132),
		PK:         rgb(232, 178, 72),
		FK:         rgb(168, 130, 240),
		Line:       rgb(116, 150, 210),
	}
}

// groupColor 用分组给表头着色，让 ER 图里"核心 / 商品 / 交易 / 营销"一眼可分。
func groupColor(t Theme, group string) types.TColor {
	if t.Dark {
		switch group {
		case "catalog":
			return rgb(42, 132, 100)
		case "trade":
			return rgb(170, 108, 48)
		case "promo":
			return rgb(116, 82, 172)
		default:
			return rgb(52, 100, 190)
		}
	}
	switch group {
	case "catalog":
		return rgb(38, 145, 104)
	case "trade":
		return rgb(208, 126, 44)
	case "promo":
		return rgb(132, 88, 198)
	default:
		return rgb(47, 111, 224)
	}
}

func groupLabel(group string) string {
	switch group {
	case "catalog":
		return "商品域"
	case "trade":
		return "交易域"
	case "promo":
		return "营销域"
	default:
		return "核心域"
	}
}

var groupKeys = []string{"core", "catalog", "trade", "promo"}
