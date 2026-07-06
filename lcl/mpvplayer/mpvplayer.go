package main

/*
#cgo linux LDFLAGS: -lmpv

#include <mpv/client.h>
#include <locale.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static mpv_handle* mpv_create_safe() {
    setlocale(LC_NUMERIC, "C");
    return mpv_create();
}

static int mpv_command_s(mpv_handle *ctx, char **args) {
    return mpv_command(ctx, (const char **)args);
}

static int mpv_command_async_s(mpv_handle *ctx, uint64_t id, char **args) {
    return mpv_command_async(ctx, id, (const char **)args);
}

static int mpv_command_ret_s(mpv_handle *ctx, char **args, mpv_node *result) {
    return mpv_command_ret(ctx, (const char **)args, result);
}

typedef struct dump_buf {
    char *data;
    size_t len;
    size_t cap;
} dump_buf;

static void dump_reserve(dump_buf *b, size_t extra) {
    size_t need = b->len + extra + 1;
    if (need <= b->cap) return;
    size_t cap = b->cap ? b->cap : 256;
    while (cap < need) cap *= 2;
    char *p = (char *)realloc(b->data, cap);
    if (!p) return;
    b->data = p;
    b->cap = cap;
}

static void dump_append(dump_buf *b, const char *s) {
    if (!s) s = "";
    size_t n = strlen(s);
    dump_reserve(b, n);
    if (!b->data) return;
    memcpy(b->data + b->len, s, n);
    b->len += n;
    b->data[b->len] = 0;
}

static void dump_appendf(dump_buf *b, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    va_list ap2;
    va_copy(ap2, ap);
    int n = vsnprintf(NULL, 0, fmt, ap);
    va_end(ap);
    if (n < 0) {
        va_end(ap2);
        return;
    }
    dump_reserve(b, (size_t)n);
    if (b->data) {
        vsnprintf(b->data + b->len, b->cap - b->len, fmt, ap2);
        b->len += (size_t)n;
    }
    va_end(ap2);
}

static void dump_string(dump_buf *b, const char *s) {
    dump_append(b, "\"");
    for (const unsigned char *p = (const unsigned char *)(s ? s : ""); *p; p++) {
        switch (*p) {
        case '\\': dump_append(b, "\\\\"); break;
        case '"': dump_append(b, "\\\""); break;
        case '\n': dump_append(b, "\\n"); break;
        case '\r': dump_append(b, "\\r"); break;
        case '\t': dump_append(b, "\\t"); break;
        default:
            if (*p < 32) dump_appendf(b, "\\u%04x", *p);
            else dump_appendf(b, "%c", *p);
        }
    }
    dump_append(b, "\"");
}

static void dump_indent(dump_buf *b, int indent) {
    for (int i = 0; i < indent; i++) dump_append(b, "  ");
}

static void dump_node_inner(dump_buf *b, mpv_node *node, int indent) {
    if (!node) {
        dump_append(b, "null");
        return;
    }
    switch (node->format) {
    case MPV_FORMAT_STRING:
        dump_string(b, node->u.string);
        break;
    case MPV_FORMAT_FLAG:
        dump_append(b, node->u.flag ? "true" : "false");
        break;
    case MPV_FORMAT_INT64:
        dump_appendf(b, "%lld", (long long)node->u.int64);
        break;
    case MPV_FORMAT_DOUBLE:
        dump_appendf(b, "%.6g", node->u.double_);
        break;
    case MPV_FORMAT_NODE_ARRAY:
        if (!node->u.list || node->u.list->num == 0) {
            dump_append(b, "[]");
            break;
        }
        dump_append(b, "[\n");
        for (int i = 0; i < node->u.list->num; i++) {
            dump_indent(b, indent + 1);
            dump_node_inner(b, &node->u.list->values[i], indent + 1);
            if (i + 1 < node->u.list->num) dump_append(b, ",");
            dump_append(b, "\n");
        }
        dump_indent(b, indent);
        dump_append(b, "]");
        break;
    case MPV_FORMAT_NODE_MAP:
        if (!node->u.list || node->u.list->num == 0) {
            dump_append(b, "{}");
            break;
        }
        dump_append(b, "{\n");
        for (int i = 0; i < node->u.list->num; i++) {
            dump_indent(b, indent + 1);
            dump_string(b, node->u.list->keys[i]);
            dump_append(b, ": ");
            dump_node_inner(b, &node->u.list->values[i], indent + 1);
            if (i + 1 < node->u.list->num) dump_append(b, ",");
            dump_append(b, "\n");
        }
        dump_indent(b, indent);
        dump_append(b, "}");
        break;
    case MPV_FORMAT_NONE:
        dump_append(b, "null");
        break;
    default:
        dump_appendf(b, "\"<format %d>\"", node->format);
        break;
    }
}

static char* mpv_dump_node_copy(mpv_node *node) {
    dump_buf b = {0};
    dump_node_inner(&b, node, 0);
    if (!b.data) return strdup("");
    return b.data;
}

static char* mpv_get_property_node_dump(mpv_handle *ctx, const char *name, int *err) {
    mpv_node node;
    int r = mpv_get_property(ctx, name, MPV_FORMAT_NODE, &node);
    if (err) *err = r;
    if (r < 0) return NULL;
    char *out = mpv_dump_node_copy(&node);
    mpv_free_node_contents(&node);
    return out;
}

static char* mpv_command_ret_dump(mpv_handle *ctx, char **args, int *err) {
    mpv_node result;
    int r = mpv_command_ret_s(ctx, args, &result);
    if (err) *err = r;
    if (r < 0) return NULL;
    char *out = mpv_dump_node_copy(&result);
    mpv_free_node_contents(&result);
    return out;
}

static char* mpv_event_dump(mpv_event *event, int *err) {
    mpv_node node;
    int r = mpv_event_to_node(&node, event);
    if (err) *err = r;
    if (r < 0) return NULL;
    char *out = mpv_dump_node_copy(&node);
    mpv_free_node_contents(&node);
    return out;
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	gtk3 "github.com/energye/energy/v3/platform/linux/gtk3/cgo"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

const (
	asyncLoadFile uint64 = 1001 + iota
	asyncAddFile
	asyncScreenshot
	asyncGetPath
	asyncSetPause
	asyncShowText
	hookOnLoad
)

const (
	seekSliderMargin   int32 = 10
	volumeSliderMargin int32 = 8
)

type observedProperty struct {
	id     uint64
	name   string
	format C.mpv_format
}

var observed = []observedProperty{
	{1, "time-pos", C.MPV_FORMAT_DOUBLE},
	{2, "duration", C.MPV_FORMAT_DOUBLE},
	{3, "pause", C.MPV_FORMAT_FLAG},
	{4, "volume", C.MPV_FORMAT_DOUBLE},
	{5, "mute", C.MPV_FORMAT_FLAG},
	{6, "speed", C.MPV_FORMAT_DOUBLE},
	{7, "filename", C.MPV_FORMAT_STRING},
	{8, "media-title", C.MPV_FORMAT_STRING},
	{9, "playlist-pos", C.MPV_FORMAT_INT64},
	{10, "playlist-count", C.MPV_FORMAT_INT64},
	{11, "chapter", C.MPV_FORMAT_INT64},
	{12, "aid", C.MPV_FORMAT_STRING},
	{13, "sid", C.MPV_FORMAT_STRING},
	{14, "vid", C.MPV_FORMAT_STRING},
	{15, "video-codec", C.MPV_FORMAT_STRING},
	{16, "audio-codec", C.MPV_FORMAT_STRING},
	{17, "hwdec-current", C.MPV_FORMAT_STRING},
	{18, "estimated-vf-fps", C.MPV_FORMAT_DOUBLE},
	{19, "cache-buffering-state", C.MPV_FORMAT_INT64},
}

var (
	ctx           *C.mpv_handle
	statsClient   *C.mpv_handle
	embedXID      uint64
	nextAsyncID   uint64 = 5000
	seeking       bool
	seekTarget    float64
	seekHoldUntil time.Time
	volSeeking    bool
)

type TMPVForm struct {
	lcl.TEngForm

	videoPanel  lcl.IPanel
	rightPanel  lcl.IPanel
	bottomPanel lcl.IPanel
	status      lcl.IStatusBar
	eventTimer  lcl.ITimer

	btnPlay   lcl.IButton
	seekPaint lcl.IPaintBox
	volPaint  lcl.IPaintBox
	lblTime   lcl.ILabel
	lblDur    lcl.ILabel
	lblFile   lcl.ILabel
	lblInfo   lcl.ILabel

	urlEdit   lcl.IEdit
	propEdit  lcl.IEdit
	valueEdit lcl.IEdit
	speedBox  lcl.IComboBox
	apiBox    lcl.IComboBox
	loopCheck lcl.ICheckBox
	logMemo   lcl.IMemo

	playPos  float64
	duration float64
	dragPos  float64
	volume   float64

	seekHover bool
	volHover  bool
	volDrag   float64

	seekSurface sliderSurface
	volSurface  sliderSurface
}

var mf TMPVForm

// https://www.bybkw.cn/post-145.html
func main() {
	libname.UseWS = "gtk3"
	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForms(&mf)

	if len(os.Args) > 1 {
		for i, p := range os.Args[1:] {
			if i == 0 {
				loadFile(p, true)
			} else {
				loadFile(p, false)
			}
		}
	}

	lcl.Application.Run()
}

func (m *TMPVForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY Video Play")
	m.SetWidth(1180)
	m.SetHeight(720)
	m.ScreenCenter()

	m.createTopBar()
	m.createRightPanel()
	m.createBottomBar()
	m.SetOnResize(func(lcl.IObject) {
		m.layoutRightPanel()
		m.invalidateSeekBar()
		m.invalidateVolumeBar()
	})

	m.videoPanel = lcl.NewPanel(m)
	m.videoPanel.SetParent(m)
	m.videoPanel.SetAlign(types.AlClient)
	m.videoPanel.SetColor(colors.ClBlack)
	m.videoPanel.SetBevelOuter(types.BvNone)

	m.status = lcl.NewStatusBar(m)
	m.status.SetParent(m)
	m.status.SetSimpleText("正在初始化 libmpv")

	if err := initMPV(m.videoPanel); err != nil {
		m.setStatus("初始化失败: " + err.Error())
		m.appendLog("init", err.Error())
		return
	}

	m.eventTimer = lcl.NewTimer(m)
	m.eventTimer.SetInterval(30)
	m.eventTimer.SetOnTimer(func(lcl.IObject) { processEvents() })
	m.eventTimer.SetEnabled(true)

	m.appendLog("mpv", fmt.Sprintf("client=%s id=%d api=0x%x xid=%d",
		mpvClientName(), mpvClientID(), uint64(C.mpv_client_api_version()), embedXID))
	m.setStatus("就绪: 打开文件、输入 URL，或从命令行传入媒体路径")
}

func (m *TMPVForm) createTopBar() {
	top := lcl.NewPanel(m)
	top.SetParent(m)
	top.SetAlign(types.AlTop)
	top.SetHeight(74)
	top.SetBevelOuter(types.BvNone)

	x := int32(8)
	button(top, x, 6, 74, "打开", openFile)
	x += 80
	button(top, x, 6, 88, "添加队列", addFile)
	x += 94
	m.btnPlay = button(top, x, 6, 66, "播放", func() { mustCommand("cycle", "pause") })
	x += 72
	button(top, x, 6, 58, "停止", func() { mustCommand("stop") })
	x += 64
	button(top, x, 6, 58, "上一个", func() { mustCommand("playlist-prev", "weak") })
	x += 64
	button(top, x, 6, 58, "下一个", func() { mustCommand("playlist-next", "weak") })
	x += 64
	button(top, x, 6, 52, "-10s", func() { mustCommand("seek", "-10", "relative") })
	x += 58
	button(top, x, 6, 52, "+10s", func() { mustCommand("seek", "10", "relative") })
	x += 58
	button(top, x, 6, 70, "截图", func() {
		check("screenshot async", mpvCommandAsync(asyncScreenshot, "screenshot-to-file", screenshotName(), "video"))
	})
	x += 76
	button(top, x, 6, 72, "逐帧", func() { mustCommand("frame-step") })
	x += 78
	button(top, x, 6, 72, "退帧", func() { mustCommand("frame-back-step") })
	x += 78

	m.loopCheck = lcl.NewCheckBox(top)
	m.loopCheck.SetParent(top)
	m.loopCheck.SetBounds(x, 9, 92, 24)
	m.loopCheck.SetCaption("循环")
	m.loopCheck.SetOnClick(func(lcl.IObject) {
		if m.loopCheck.Checked() {
			check("set loop-file", mpvSetPropertyString("loop-file", "inf"))
		} else {
			check("set loop-file", mpvSetPropertyString("loop-file", "no"))
		}
	})

	lb := lcl.NewLabel(top)
	lb.SetParent(top)
	lb.SetBounds(8, 46, 34, 20)
	lb.SetCaption("URL")

	m.urlEdit = lcl.NewEdit(top)
	m.urlEdit.SetParent(top)
	m.urlEdit.SetBounds(44, 40, 560, 28)
	m.urlEdit.SetTextHint("https://... 或 /path/to/video.mp4")

	button(top, 612, 40, 70, "播放URL", func() {
		p := strings.TrimSpace(m.urlEdit.Text())
		if p != "" {
			loadFile(p, true)
		}
	})
	button(top, 690, 40, 80, "添加URL", func() {
		p := strings.TrimSpace(m.urlEdit.Text())
		if p != "" {
			loadFile(p, false)
		}
	})

	lbSpeed := lcl.NewLabel(top)
	lbSpeed.SetParent(top)
	lbSpeed.SetBounds(790, 46, 34, 20)
	lbSpeed.SetCaption("速度")

	m.speedBox = lcl.NewComboBox(top)
	m.speedBox.SetParent(top)
	m.speedBox.SetBounds(826, 40, 86, 28)
	m.speedBox.SetStyle(types.CsDropDownList)
	for _, s := range []string{"0.5", "0.75", "1.0", "1.25", "1.5", "2.0"} {
		m.speedBox.Items().Add(s)
	}
	m.speedBox.SetItemIndex(2)
	m.speedBox.SetOnChange(func(lcl.IObject) {
		if m.speedBox.ItemIndex() >= 0 {
			check("set speed", mpvSetPropertyDouble("speed", parseFloat(m.speedBox.Items().Strings(m.speedBox.ItemIndex()), 1)))
		}
	})

	button(top, 922, 40, 58, "静音", func() { mustCommand("cycle", "mute") })
	button(top, 986, 40, 78, "保存进度", func() { mustCommand("write-watch-later-config") })
}

func (m *TMPVForm) createRightPanel() {
	m.rightPanel = lcl.NewPanel(m)
	m.rightPanel.SetParent(m)
	m.rightPanel.SetAlign(types.AlRight)
	m.rightPanel.SetWidth(330)
	m.rightPanel.SetBevelOuter(types.BvLowered)

	title := lcl.NewLabel(m.rightPanel)
	title.SetParent(m.rightPanel)
	title.SetBounds(10, 10, 200, 20)
	title.SetCaption("libmpv API 面板")

	m.lblFile = lcl.NewLabel(m.rightPanel)
	m.lblFile.SetParent(m.rightPanel)
	m.lblFile.SetBounds(10, 36, 305, 42)
	m.lblFile.SetCaption("未加载")

	m.lblInfo = lcl.NewLabel(m.rightPanel)
	m.lblInfo.SetParent(m.rightPanel)
	m.lblInfo.SetBounds(10, 80, 305, 42)
	m.lblInfo.SetCaption("视频/音频信息等待加载")

	lbProp := lcl.NewLabel(m.rightPanel)
	lbProp.SetParent(m.rightPanel)
	lbProp.SetBounds(10, 128, 80, 20)
	lbProp.SetCaption("属性名")

	m.propEdit = lcl.NewEdit(m.rightPanel)
	m.propEdit.SetParent(m.rightPanel)
	m.propEdit.SetBounds(76, 124, 230, 26)
	m.propEdit.SetText("pause")

	lbVal := lcl.NewLabel(m.rightPanel)
	lbVal.SetParent(m.rightPanel)
	lbVal.SetBounds(10, 158, 80, 20)
	lbVal.SetCaption("属性值")

	m.valueEdit = lcl.NewEdit(m.rightPanel)
	m.valueEdit.SetParent(m.rightPanel)
	m.valueEdit.SetBounds(76, 154, 230, 26)
	m.valueEdit.SetText("yes")

	button(m.rightPanel, 10, 188, 88, "Get 字符串", func() {
		name := strings.TrimSpace(m.propEdit.Text())
		if name == "" {
			return
		}
		v, err := mpvGetPropertyString(name)
		if err != nil {
			m.appendLog("get_property", err.Error())
			return
		}
		m.valueEdit.SetText(v)
		m.appendLog("get_property", name+" = "+v)
	})
	button(m.rightPanel, 106, 188, 88, "Set 字符串", func() {
		name := strings.TrimSpace(m.propEdit.Text())
		if name != "" {
			check("set_property_string", mpvSetPropertyString(name, m.valueEdit.Text()))
		}
	})
	button(m.rightPanel, 202, 188, 104, "Set 异步", func() {
		name := strings.TrimSpace(m.propEdit.Text())
		if name != "" {
			check("set_property_async", mpvSetPropertyStringAsync(nextID(), name, m.valueEdit.Text()))
		}
	})

	m.apiBox = lcl.NewComboBox(m.rightPanel)
	m.apiBox.SetParent(m.rightPanel)
	m.apiBox.SetBounds(10, 226, 296, 28)
	m.apiBox.SetStyle(types.CsDropDownList)
	for _, s := range []string{
		"Node: playlist / track-list / params",
		"CommandRet: get_property media-title",
		"CommandAsync: get_property path",
		"OSD String: time-pos / percent-pos",
		"Client API: create_client / weak_client",
		"EventToNode: dump next event",
		"Tracks: cycle audio/sub/video",
		"Video filters: rotate / deinterlace / reset",
		"Playlist: shuffle / clear",
	} {
		m.apiBox.Items().Add(s)
	}
	m.apiBox.SetItemIndex(0)

	button(m.rightPanel, 10, 262, 92, "运行 API", runSelectedAPI)
	button(m.rightPanel, 108, 262, 92, "清空日志", func() { m.logMemo.Lines().Clear() })
	button(m.rightPanel, 206, 262, 100, "显示 OSD", func() {
		check("show-text async", mpvCommandAsync(asyncShowText, "show-text",
			"${filename}\\n${time-pos}/${duration}\\n${video-codec} ${audio-codec}", "3000"))
	})

	m.logMemo = lcl.NewMemo(m.rightPanel)
	m.logMemo.SetParent(m.rightPanel)
	m.logMemo.SetBounds(10, 302, 296, 340)
	m.logMemo.SetScrollBars(types.SsVertical)
	m.logMemo.SetReadOnly(true)
	m.layoutRightPanel()
}

func (m *TMPVForm) layoutRightPanel() {
	if m.rightPanel == nil || m.logMemo == nil {
		return
	}

	w := m.rightPanel.Width()
	h := m.rightPanel.Height()
	logTop := int32(302)
	bottomPad := int32(10)
	logW := w - 20
	logH := h - logTop - bottomPad
	if logW < 120 {
		logW = 120
	}
	if logH < 80 {
		logH = 80
	}
	m.logMemo.SetBounds(10, logTop, logW, logH)
}

func (m *TMPVForm) createBottomBar() {
	m.bottomPanel = lcl.NewPanel(m)
	m.bottomPanel.SetParent(m)
	m.bottomPanel.SetAlign(types.AlBottom)
	m.bottomPanel.SetHeight(96)
	m.bottomPanel.SetBevelOuter(types.BvLowered)

	m.lblTime = lcl.NewLabel(m.bottomPanel)
	m.lblTime.SetParent(m.bottomPanel)
	m.lblTime.SetBounds(8, 14, 68, 20)
	m.lblTime.SetCaption("00:00")

	m.seekPaint = lcl.NewPaintBox(m.bottomPanel)
	m.seekPaint.SetParent(m.bottomPanel)
	m.seekPaint.SetBounds(78, 6, 640, 34)
	m.seekPaint.SetOnPaint(func(lcl.IObject) {
		m.paintSeekBar()
	})
	m.seekPaint.SetOnMouseEnter(func(lcl.IObject) {
		m.seekHover = true
		m.seekPaint.Invalidate()
	})
	m.seekPaint.SetOnMouseLeave(func(lcl.IObject) {
		if !seeking {
			m.seekHover = false
			m.seekPaint.Invalidate()
		}
	})
	m.seekPaint.SetOnMouseDown(func(_ lcl.IObject, _ types.TMouseButton, _ types.TShiftState, x int32, _ int32) {
		seeking = true
		m.dragPos = m.seekValueFromX(x)
		seekTarget = m.dragPos
		m.lblTime.SetCaption(formatTime(m.dragPos))
		m.seekPaint.Invalidate()
	})
	m.seekPaint.SetOnMouseMove(func(_ lcl.IObject, _ types.TShiftState, x int32, _ int32) {
		if seeking {
			m.dragPos = m.seekValueFromX(x)
			seekTarget = m.dragPos
			m.lblTime.SetCaption(formatTime(m.dragPos))
			m.seekPaint.Invalidate()
		}
	})
	m.seekPaint.SetOnMouseUp(func(_ lcl.IObject, _ types.TMouseButton, _ types.TShiftState, x int32, _ int32) {
		m.dragPos = m.seekValueFromX(x)
		seekTarget = m.dragPos
		m.playPos = seekTarget
		seekHoldUntil = time.Now().Add(900 * time.Millisecond)
		m.lblTime.SetCaption(formatTime(seekTarget))
		if ctx != nil {
			check("seek absolute", mpvSetPropertyDouble("time-pos", seekTarget))
		}
		seeking = false
		m.seekPaint.Invalidate()
	})

	m.lblDur = lcl.NewLabel(m.bottomPanel)
	m.lblDur.SetParent(m.bottomPanel)
	m.lblDur.SetBounds(724, 14, 74, 20)
	m.lblDur.SetCaption("00:00")

	lbVol := lcl.NewLabel(m.bottomPanel)
	lbVol.SetParent(m.bottomPanel)
	lbVol.SetBounds(8, 58, 44, 20)
	lbVol.SetCaption("音量")

	m.volume = 80
	m.volPaint = lcl.NewPaintBox(m.bottomPanel)
	m.volPaint.SetParent(m.bottomPanel)
	m.volPaint.SetBounds(52, 50, 150, 34)
	m.volPaint.SetOnPaint(func(lcl.IObject) {
		m.paintVolumeBar()
	})
	m.volPaint.SetOnMouseEnter(func(lcl.IObject) {
		m.volHover = true
		m.volPaint.Invalidate()
	})
	m.volPaint.SetOnMouseLeave(func(lcl.IObject) {
		if !volSeeking {
			m.volHover = false
			m.volPaint.Invalidate()
		}
	})
	m.volPaint.SetOnMouseDown(func(_ lcl.IObject, _ types.TMouseButton, _ types.TShiftState, x int32, _ int32) {
		volSeeking = true
		m.volDrag = m.volumeValueFromX(x)
		m.volume = m.volDrag
		m.volPaint.Invalidate()
		if ctx != nil {
			check("set volume", mpvSetPropertyDouble("volume", m.volDrag))
		}
	})
	m.volPaint.SetOnMouseMove(func(_ lcl.IObject, _ types.TShiftState, x int32, _ int32) {
		if volSeeking {
			m.volDrag = m.volumeValueFromX(x)
			m.volume = m.volDrag
			m.volPaint.Invalidate()
			if ctx != nil {
				check("set volume", mpvSetPropertyDouble("volume", m.volDrag))
			}
		}
	})
	m.volPaint.SetOnMouseUp(func(_ lcl.IObject, _ types.TMouseButton, _ types.TShiftState, x int32, _ int32) {
		m.volDrag = m.volumeValueFromX(x)
		m.volume = m.volDrag
		volSeeking = false
		m.volPaint.Invalidate()
		if ctx != nil {
			check("set volume", mpvSetPropertyDouble("volume", m.volDrag))
		}
	})

	button(m.bottomPanel, 216, 54, 64, "字幕", func() { mustCommand("cycle", "sid") })
	button(m.bottomPanel, 286, 54, 64, "音轨", func() { mustCommand("cycle", "aid") })
	button(m.bottomPanel, 356, 54, 70, "视频轨", func() { mustCommand("cycle", "vid") })
	button(m.bottomPanel, 432, 54, 78, "字幕文件", openSubtitle)
	button(m.bottomPanel, 516, 54, 82, "绝对50%", func() { check("percent-pos", mpvSetPropertyDouble("percent-pos", 50)) })
	button(m.bottomPanel, 604, 54, 86, "全屏切换", func() { mustCommand("cycle", "fullscreen") })
}

func initMPV(panel lcl.IPanel) error {
	ctx = C.mpv_create_safe()
	if ctx == nil {
		return fmt.Errorf("mpv_create returned nil")
	}

	for _, opt := range [][2]string{
		{"osc", "no"},
		{"keep-open", "yes"},
		{"idle", "yes"},
		{"force-window", "yes"},
		{"input-default-bindings", "yes"},
		{"input-vo-keyboard", "yes"},
		{"terminal", "yes"},
		{"hwdec", "auto-safe"},
		{"cache", "yes"},
		{"demuxer-max-bytes", "128MiB"},
	} {
		if err := mpvSetOptionString(opt[0], opt[1]); err != nil {
			return fmt.Errorf("set option %s: %w", opt[0], err)
		}
	}
	check("typed option volume", mpvSetOptionDouble("volume", 80))

	gtk3widget := lcl.PlatformHandle(panel.Handle()).Gtk3Widget()
	widget := gtk3.AsWidget(unsafe.Pointer(gtk3widget))
	embedXID = uint64(gtk3.WindowX11ID(widget))
	if embedXID != 0 {
		if err := mpvSetOptionString("wid", fmt.Sprintf("%d", embedXID)); err != nil {
			return fmt.Errorf("set wid: %w", err)
		}
	}

	if err := mpvErr(int(C.mpv_initialize(ctx))); err != nil {
		return err
	}

	check("request logs", withCString("info", func(level *C.char) error {
		return mpvErr(int(C.mpv_request_log_messages(ctx, level)))
	}))
	check("request events", mpvErr(int(C.mpv_request_event(ctx, C.MPV_EVENT_LOG_MESSAGE, 1))))

	for _, p := range observed {
		if err := observeProperty(p); err != nil {
			mf.appendLog("observe", p.name+": "+err.Error())
		}
	}

	statsClientName := C.CString("lcl-stats-client")
	statsClient = C.mpv_create_client(ctx, statsClientName)
	C.free(unsafe.Pointer(statsClientName))
	if statsClient == nil {
		mf.appendLog("client", "create_client returned nil")
	}

	check("hook on_load", withCString("on_load", func(name *C.char) error {
		return mpvErr(int(C.mpv_hook_add(ctx, C.uint64_t(hookOnLoad), name, 50)))
	}))
	return nil
}

func observeProperty(p observedProperty) error {
	name := C.CString(p.name)
	defer C.free(unsafe.Pointer(name))
	return mpvErr(int(C.mpv_observe_property(ctx, C.uint64_t(p.id), name, p.format)))
}

func processEvents() {
	if ctx == nil {
		return
	}
	for {
		ev := C.mpv_wait_event(ctx, 0)
		if ev == nil || ev.event_id == C.MPV_EVENT_NONE {
			break
		}

		switch ev.event_id {
		case C.MPV_EVENT_LOG_MESSAGE:
			msg := (*C.mpv_event_log_message)(ev.data)
			if msg != nil {
				text := strings.TrimSpace(C.GoString(msg.text))
				if text != "" {
					mf.appendLog(C.GoString(msg.prefix)+"/"+C.GoString(msg.level), text)
				}
			}
		case C.MPV_EVENT_START_FILE:
			mf.setStatus("开始加载媒体")
		case C.MPV_EVENT_FILE_LOADED:
			onFileLoaded()
		case C.MPV_EVENT_END_FILE:
			onEndFile(ev)
		case C.MPV_EVENT_PROPERTY_CHANGE:
			onPropertyChange((*C.mpv_event_property)(ev.data))
		case C.MPV_EVENT_COMMAND_REPLY:
			onCommandReply(ev)
		case C.MPV_EVENT_SET_PROPERTY_REPLY:
			if ev.error < 0 {
				mf.appendLog("set_property_reply", fmt.Sprintf("id=%d %s", uint64(ev.reply_userdata), mpvErrorString(int(ev.error))))
			} else {
				mf.appendLog("set_property_reply", fmt.Sprintf("id=%d ok", uint64(ev.reply_userdata)))
			}
		case C.MPV_EVENT_HOOK:
			h := (*C.mpv_event_hook)(ev.data)
			if h != nil {
				mf.appendLog("hook", fmt.Sprintf("%s id=%d", C.GoString(h.name), uint64(h.id)))
				C.mpv_hook_continue(ctx, h.id)
			}
		case C.MPV_EVENT_SHUTDOWN:
			mf.appendLog("event", "shutdown")
			ctx = nil
			return
		default:
			name := C.GoString(C.mpv_event_name(ev.event_id))
			if ev.error < 0 {
				mf.appendLog("event", name+": "+mpvErrorString(int(ev.error)))
			}
		}
	}
}

func onFileLoaded() {
	title, _ := mpvGetPropertyString("media-title")
	filename, _ := mpvGetPropertyString("filename")
	if title == "" {
		title = filename
	}
	mf.lblFile.SetCaption(trimForLabel(title, 90))
	mf.setStatus("加载完成: " + title)
	refreshMediaInfo()
}

func onEndFile(ev *C.mpv_event) {
	end := (*C.mpv_event_end_file)(ev.data)
	if end == nil {
		mf.setStatus("播放结束")
		return
	}
	msg := fmt.Sprintf("播放结束 reason=%d", int(end.reason))
	if end.error < 0 {
		msg += " error=" + mpvErrorString(int(end.error))
	}
	mf.setStatus(msg)
	mf.appendLog("end_file", msg)
}

func onCommandReply(ev *C.mpv_event) {
	cmd := (*C.mpv_event_command)(ev.data)
	if ev.error < 0 {
		mf.appendLog("command_reply", fmt.Sprintf("id=%d %s", uint64(ev.reply_userdata), mpvErrorString(int(ev.error))))
		return
	}
	if cmd == nil {
		mf.appendLog("command_reply", fmt.Sprintf("id=%d ok", uint64(ev.reply_userdata)))
		return
	}
	out := dumpNode(&cmd.result)
	if strings.TrimSpace(out) == "null" || out == "" {
		out = "ok"
	}
	mf.appendLog("command_reply", fmt.Sprintf("id=%d %s", uint64(ev.reply_userdata), out))
}

func onPropertyChange(prop *C.mpv_event_property) {
	if prop == nil || prop.name == nil {
		return
	}
	name := C.GoString(prop.name)
	if prop.format == C.MPV_FORMAT_NONE || prop.data == nil {
		return
	}

	switch prop.format {
	case C.MPV_FORMAT_DOUBLE:
		v := float64(*(*C.double)(prop.data))
		switch name {
		case "time-pos":
			if seeking {
				mf.lblTime.SetCaption(formatTime(seekTarget))
				mf.invalidateSeekBar()
				return
			}
			if !seekHoldUntil.IsZero() {
				if time.Now().Before(seekHoldUntil) && absFloat(v-seekTarget) > 1.25 {
					mf.playPos = seekTarget
					mf.lblTime.SetCaption(formatTime(seekTarget))
					mf.invalidateSeekBar()
					return
				}
				seekHoldUntil = time.Time{}
			}
			mf.playPos = v
			mf.lblTime.SetCaption(formatTime(v))
			mf.invalidateSeekBar()
		case "duration":
			if v > 0 {
				mf.duration = v
			}
			mf.lblDur.SetCaption(formatTime(v))
			mf.invalidateSeekBar()
		case "volume":
			if !volSeeking {
				mf.volume = v
				mf.invalidateVolumeBar()
			}
		case "speed":
			mf.setStatus(fmt.Sprintf("速度 %.2fx", v))
		case "estimated-vf-fps":
			refreshMediaInfo()
		}
	case C.MPV_FORMAT_INT64:
		v := int64(*(*C.int64_t)(prop.data))
		switch name {
		case "playlist-pos", "playlist-count", "chapter", "cache-buffering-state":
			mf.appendLog("property", fmt.Sprintf("%s=%d", name, v))
		}
	case C.MPV_FORMAT_FLAG:
		v := int(*(*C.int)(prop.data)) != 0
		switch name {
		case "pause":
			if v {
				mf.btnPlay.SetCaption("播放")
			} else {
				mf.btnPlay.SetCaption("暂停")
			}
		case "mute":
			if v {
				mf.setStatus("已静音")
			}
		}
	case C.MPV_FORMAT_STRING:
		v := C.GoString(*(**C.char)(prop.data))
		switch name {
		case "filename", "media-title":
			if v != "" {
				mf.lblFile.SetCaption(trimForLabel(v, 90))
			}
		case "aid", "sid", "vid", "video-codec", "audio-codec", "hwdec-current":
			refreshMediaInfo()
		}
	}
}

func runSelectedAPI() {
	idx := mf.apiBox.ItemIndex()
	switch idx {
	case 0:
		for _, prop := range []string{"playlist", "track-list", "video-params", "audio-params"} {
			out, err := mpvGetPropertyNodeDump(prop)
			if err != nil {
				mf.appendLog("node "+prop, err.Error())
			} else {
				mf.appendLog("node "+prop, out)
			}
		}
	case 1:
		out, err := mpvCommandRetDump("get_property", "media-title")
		if err != nil {
			mf.appendLog("command_ret", err.Error())
		} else {
			mf.appendLog("command_ret", out)
		}
	case 2:
		check("command_async get_property", mpvCommandAsync(asyncGetPath, "get_property", "path"))
	case 3:
		for _, prop := range []string{"time-pos", "duration", "percent-pos", "volume"} {
			out, err := mpvGetOSDString(prop)
			if err != nil {
				mf.appendLog("osd", prop+": "+err.Error())
			} else {
				mf.appendLog("osd", prop+" = "+out)
			}
		}
	case 4:
		demoClients()
	case 5:
		dumpNextEvent()
	case 6:
		mustCommand("cycle", "aid")
		mustCommand("cycle", "sid")
		mustCommand("cycle", "vid")
	case 7:
		mustCommand("vf", "toggle", "lavfi=[transpose=clock]")
		mustCommand("cycle", "deinterlace")
		mf.appendLog("filters", "toggle transpose + cycle deinterlace; 用 reset 过滤器可恢复")
	case 8:
		mustCommand("playlist-shuffle")
		mf.appendLog("playlist", "已调用 playlist-shuffle；清空队列可执行 playlist-clear")
	}
}

func dumpNextEvent() {
	ev := C.mpv_wait_event(ctx, 0)
	if ev == nil || ev.event_id == C.MPV_EVENT_NONE {
		mf.appendLog("event_to_node", "当前事件队列为空")
		return
	}
	var err C.int
	s := C.mpv_event_dump(ev, &err)
	if err < 0 {
		mf.appendLog("event_to_node", mpvErrorString(int(err)))
		return
	}
	defer C.free(unsafe.Pointer(s))
	mf.appendLog("event_to_node", C.GoString(s))
}

func demoClients() {
	if statsClient != nil {
		mf.appendLog("client", fmt.Sprintf("stats client id=%d name=%s",
			int64(C.mpv_client_id(statsClient)), C.GoString(C.mpv_client_name(statsClient))))
	}
	weakName := C.CString("lcl-weak-client")
	weak := C.mpv_create_weak_client(ctx, weakName)
	C.free(unsafe.Pointer(weakName))
	if weak == nil {
		mf.appendLog("client", "create_weak_client returned nil")
		return
	}
	mf.appendLog("client", fmt.Sprintf("weak client id=%d name=%s",
		int64(C.mpv_client_id(weak)), C.GoString(C.mpv_client_name(weak))))
	C.mpv_destroy(weak)
}

func refreshMediaInfo() {
	video, _ := mpvGetPropertyString("video-codec")
	audio, _ := mpvGetPropertyString("audio-codec")
	hwdec, _ := mpvGetPropertyString("hwdec-current")
	fps, _ := mpvGetPropertyDouble("estimated-vf-fps")
	if hwdec == "" {
		hwdec = "no"
	}
	info := fmt.Sprintf("V:%s  A:%s  FPS:%.2f  HW:%s", blank(video), blank(audio), fps, hwdec)
	mf.lblInfo.SetCaption(trimForLabel(info, 90))
}

func openFile() {
	d := lcl.NewOpenDialog(&mf)
	d.SetFilter("媒体文件|*.mp4;*.mkv;*.avi;*.mov;*.webm;*.flv;*.mp3;*.flac;*.m4a;*.wav|所有文件|*.*")
	if d.Execute() {
		loadFile(d.FileName(), true)
	}
}

func addFile() {
	d := lcl.NewOpenDialog(&mf)
	d.SetFilter("媒体文件|*.mp4;*.mkv;*.avi;*.mov;*.webm;*.flv;*.mp3;*.flac;*.m4a;*.wav|所有文件|*.*")
	if d.Execute() {
		loadFile(d.FileName(), false)
	}
}

func openSubtitle() {
	d := lcl.NewOpenDialog(&mf)
	d.SetFilter("字幕|*.srt;*.ass;*.ssa;*.vtt|所有文件|*.*")
	if d.Execute() {
		mustCommand("sub-add", d.FileName(), "select")
	}
}

func loadFile(path string, replace bool) {
	if ctx == nil {
		return
	}
	mode := "replace"
	id := asyncLoadFile
	if !replace {
		mode = "append-play"
		id = asyncAddFile
	}
	if err := mpvCommandAsync(id, "loadfile", path, mode); err != nil {
		mf.appendLog("loadfile", err.Error())
		mf.setStatus("加载失败: " + err.Error())
		return
	}
	mf.setStatus("加载: " + path)
}

func mustCommand(args ...string) {
	if err := mpvCommand(args...); err != nil {
		mf.appendLog("command", strings.Join(args, " ")+" -> "+err.Error())
		mf.setStatus(err.Error())
	}
}

func check(label string, err error) {
	if err != nil {
		mf.appendLog(label, err.Error())
	}
}

func mpvCommand(args ...string) error {
	if ctx == nil {
		return fmt.Errorf("mpv is not initialized")
	}
	cargs := makeCArgs(args)
	defer freeCArgs(cargs)
	return mpvErr(int(C.mpv_command_s(ctx, &cargs[0])))
}

func mpvCommandAsync(id uint64, args ...string) error {
	if ctx == nil {
		return fmt.Errorf("mpv is not initialized")
	}
	cargs := makeCArgs(args)
	defer freeCArgs(cargs)
	return mpvErr(int(C.mpv_command_async_s(ctx, C.uint64_t(id), &cargs[0])))
}

func mpvCommandString(command string) error {
	return withCString(command, func(c *C.char) error {
		return mpvErr(int(C.mpv_command_string(ctx, c)))
	})
}

func mpvCommandRetDump(args ...string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("mpv is not initialized")
	}
	cargs := makeCArgs(args)
	defer freeCArgs(cargs)
	var err C.int
	s := C.mpv_command_ret_dump(ctx, &cargs[0], &err)
	if err < 0 {
		return "", mpvErr(int(err))
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s), nil
}

func mpvSetOptionString(name, value string) error {
	return withCString2(name, value, func(n, v *C.char) error {
		return mpvErr(int(C.mpv_set_option_string(ctx, n, v)))
	})
}

func mpvSetOptionDouble(name string, value float64) error {
	return withCString(name, func(n *C.char) error {
		v := C.double(value)
		return mpvErr(int(C.mpv_set_option(ctx, n, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&v))))
	})
}

func mpvSetPropertyString(name, value string) error {
	return withCString2(name, value, func(n, v *C.char) error {
		return mpvErr(int(C.mpv_set_property_string(ctx, n, v)))
	})
}

func mpvSetPropertyStringAsync(id uint64, name, value string) error {
	return withCString2(name, value, func(n, v *C.char) error {
		return mpvErr(int(C.mpv_set_property_async(ctx, C.uint64_t(id), n, C.MPV_FORMAT_STRING, unsafe.Pointer(&v))))
	})
}

func mpvSetPropertyDouble(name string, value float64) error {
	return withCString(name, func(n *C.char) error {
		v := C.double(value)
		return mpvErr(int(C.mpv_set_property(ctx, n, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&v))))
	})
}

func mpvSetPropertyFlag(name string, value bool) error {
	return withCString(name, func(n *C.char) error {
		v := C.int(0)
		if value {
			v = 1
		}
		return mpvErr(int(C.mpv_set_property(ctx, n, C.MPV_FORMAT_FLAG, unsafe.Pointer(&v))))
	})
}

func mpvGetPropertyString(name string) (string, error) {
	var ret *C.char
	err := withCString(name, func(n *C.char) error {
		ret = C.mpv_get_property_string(ctx, n)
		if ret == nil {
			return fmt.Errorf("%s: %s", name, mpvErrorString(int(C.MPV_ERROR_PROPERTY_UNAVAILABLE)))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	defer C.mpv_free(unsafe.Pointer(ret))
	return C.GoString(ret), nil
}

func mpvGetPropertyDouble(name string) (float64, error) {
	var v C.double
	err := withCString(name, func(n *C.char) error {
		return mpvErr(int(C.mpv_get_property(ctx, n, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&v))))
	})
	return float64(v), err
}

func mpvGetOSDString(name string) (string, error) {
	var ret *C.char
	err := withCString(name, func(n *C.char) error {
		ret = C.mpv_get_property_osd_string(ctx, n)
		if ret == nil {
			return fmt.Errorf("%s: osd string unavailable", name)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	defer C.mpv_free(unsafe.Pointer(ret))
	return C.GoString(ret), nil
}

func mpvGetPropertyNodeDump(name string) (string, error) {
	var err C.int
	var s *C.char
	e := withCString(name, func(n *C.char) error {
		s = C.mpv_get_property_node_dump(ctx, n, &err)
		if err < 0 {
			return mpvErr(int(err))
		}
		return nil
	})
	if e != nil {
		return "", e
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s), nil
}

func dumpNode(node *C.mpv_node) string {
	s := C.mpv_dump_node_copy(node)
	if s == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s)
}

func mpvClientName() string {
	if ctx == nil {
		return ""
	}
	return C.GoString(C.mpv_client_name(ctx))
}

func mpvClientID() int64 {
	if ctx == nil {
		return 0
	}
	return int64(C.mpv_client_id(ctx))
}

func mpvErr(code int) error {
	if code >= 0 {
		return nil
	}
	return fmt.Errorf(mpvErrorString(code))
}

func mpvErrorString(code int) string {
	return C.GoString(C.mpv_error_string(C.int(code)))
}

func makeCArgs(args []string) []*C.char {
	if len(args) == 0 {
		args = []string{""}
	}
	cargs := make([]*C.char, len(args)+1)
	for i, a := range args {
		cargs[i] = C.CString(a)
	}
	return cargs
}

func freeCArgs(args []*C.char) {
	for _, a := range args {
		if a != nil {
			C.free(unsafe.Pointer(a))
		}
	}
}

func withCString(s string, fn func(*C.char) error) error {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return fn(cs)
}

func withCString2(a, b string, fn func(*C.char, *C.char) error) error {
	ca := C.CString(a)
	cb := C.CString(b)
	defer C.free(unsafe.Pointer(ca))
	defer C.free(unsafe.Pointer(cb))
	return fn(ca, cb)
}

func nextID() uint64 {
	nextAsyncID++
	return nextAsyncID
}

func (m *TMPVForm) appendLog(prefix, text string) {
	if m.logMemo == nil {
		return
	}
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		m.logMemo.Lines().Add(fmt.Sprintf("[%s] %s", prefix, line))
	}
}

func (m *TMPVForm) setStatus(s string) {
	if m.status != nil {
		m.status.SetSimpleText(s)
	}
}

func (m *TMPVForm) paintSeekBar() {
	if m.seekPaint == nil {
		return
	}
	pos := m.playPos
	if seeking {
		pos = m.dragPos
	}
	m.paintSlider(m.seekPaint, pos, m.duration, seeking || m.seekHover, sliderStyle{
		TrackColor: colors.ClLightgray,
		FillColor:  colors.ClDodgerblue,
		KnobColor:  colors.ClWhite,
		EdgeColor:  colors.ClDodgerblue,
		Margin:     seekSliderMargin,
		TrackHalf:  4,
		Knob:       5,
		KnobActive: 7,
	}, &m.seekSurface)
}

func (m *TMPVForm) paintVolumeBar() {
	if m.volPaint == nil {
		return
	}
	value := m.volume
	if volSeeking {
		value = m.volDrag
	}
	m.paintSlider(m.volPaint, value, 130, volSeeking || m.volHover, sliderStyle{
		TrackColor: colors.ClLightgray,
		FillColor:  colors.ClLimegreen,
		KnobColor:  colors.ClWhite,
		EdgeColor:  colors.ClForestgreen,
		Margin:     volumeSliderMargin,
		TrackHalf:  4,
		Knob:       5,
		KnobActive: 7,
	}, &m.volSurface)
}

func (m *TMPVForm) seekValueFromX(x int32) float64 {
	if m.seekPaint == nil || m.duration <= 0 {
		return 0
	}
	left := seekSliderMargin
	right := m.seekPaint.Width() - seekSliderMargin
	if right <= left {
		return 0
	}
	if x < left {
		x = left
	}
	if x > right {
		x = right
	}
	return float64(x-left) / float64(right-left) * m.duration
}

func (m *TMPVForm) volumeValueFromX(x int32) float64 {
	if m.volPaint == nil {
		return 0
	}
	left := volumeSliderMargin
	right := m.volPaint.Width() - volumeSliderMargin
	if right <= left {
		return 0
	}
	if x < left {
		x = left
	}
	if x > right {
		x = right
	}
	return float64(x-left) / float64(right-left) * 130
}

func (m *TMPVForm) invalidateSeekBar() {
	if m.seekPaint != nil {
		m.seekPaint.Invalidate()
	}
}

func (m *TMPVForm) invalidateVolumeBar() {
	if m.volPaint != nil {
		m.volPaint.Invalidate()
	}
}

type sliderStyle struct {
	TrackColor types.TColor
	FillColor  types.TColor
	KnobColor  types.TColor
	EdgeColor  types.TColor
	Margin     int32
	TrackHalf  int32
	Knob       int32
	KnobActive int32
}

type sliderSurface struct {
	img    lcl.ILazIntfImage
	bitmap lcl.IBitmap
}

func (s *sliderSurface) ensure(w, h int32) {
	if s.img == nil || !s.img.IsValid() {
		s.img = lcl.NewLazIntfImageWithIntX2RIQFlags(0, 0, types.NewSet(types.RiqfRGB, types.RiqfAlpha))
	}
	if s.bitmap == nil || !s.bitmap.IsValid() {
		s.bitmap = lcl.NewBitmap()
		s.bitmap.SetPixelFormat(types.Pf32bit)
	}
	if s.img.Width() != w || s.img.Height() != h {
		s.img.SetSize(w, h)
	}
	if s.bitmap.Width() != w || s.bitmap.Height() != h {
		s.bitmap.SetSize(w, h)
	}
}

func (m *TMPVForm) paintSlider(paint lcl.IPaintBox, value, max float64, active bool, style sliderStyle, surface *sliderSurface) {
	canvas := paint.Canvas()
	w := paint.Width()
	h := paint.Height()
	if w <= 0 || h <= 0 {
		return
	}

	left := style.Margin
	right := w - style.Margin
	if right <= left {
		return
	}
	if value < 0 {
		value = 0
	}
	if max > 0 && value > max {
		value = max
	}
	ratio := 0.0
	if max > 0 {
		ratio = value / max
	}

	centerY := float64(h) / 2
	trackTop := centerY - float64(style.TrackHalf)
	trackBottom := centerY + float64(style.TrackHalf)
	progressX := left + int32(ratio*float64(right-left))
	if progressX < left {
		progressX = left
	}
	if progressX > right {
		progressX = right
	}

	if surface == nil {
		return
	}
	surface.ensure(w, h)

	knob := style.Knob
	if active {
		knob = style.KnobActive
	}
	knobHalfW := float64(knob)
	knobHalfH := float64(9)
	if active {
		knobHalfH = 11
	}

	scale := int32(4)
	bg := sliderRGB{246, 247, 249}
	track := colorToSliderRGB(style.TrackColor)
	fill := colorToSliderRGB(style.FillColor)
	edge := colorToSliderRGB(style.EdgeColor)
	knobRGB := colorToSliderRGB(style.KnobColor)
	shadow := sliderRGB{148, 163, 184}

	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			r, g, b := bg.r, bg.g, bg.b
			r, g, b = blendSliderPixel(r, g, b, shadow, roundedRectCoverage(x, y, scale,
				float64(progressX)-knobHalfW+1, centerY-knobHalfH+2,
				float64(progressX)+knobHalfW+1, centerY+knobHalfH+2, 4), 0.22)
			r, g, b = blendSliderPixel(r, g, b, track, roundedRectCoverage(x, y, scale,
				float64(left), trackTop, float64(right), trackBottom, float64(style.TrackHalf)), 1)
			if progressX > left {
				r, g, b = blendSliderPixel(r, g, b, fill, roundedRectCoverage(x, y, scale,
					float64(left), trackTop, float64(progressX), trackBottom, float64(style.TrackHalf)), 1)
			}
			r, g, b = blendSliderPixel(r, g, b, edge, roundedRectCoverage(x, y, scale,
				float64(progressX)-knobHalfW, centerY-knobHalfH,
				float64(progressX)+knobHalfW, centerY+knobHalfH, 4), 1)
			r, g, b = blendSliderPixel(r, g, b, knobRGB, roundedRectCoverage(x, y, scale,
				float64(progressX)-knobHalfW+2, centerY-knobHalfH+2,
				float64(progressX)+knobHalfW-2, centerY+knobHalfH-2, 3), 1)
			surface.img.SetColors(x, y, lcl.TFPColor{
				Red:   uint16(r) << 8,
				Green: uint16(g) << 8,
				Blue:  uint16(b) << 8,
				Alpha: 0xffff,
			})
		}
	}
	surface.bitmap.LoadFromIntfImage(surface.img)
	canvas.DrawWithIntX2Graphic(0, 0, surface.bitmap)
}

type sliderRGB struct {
	r byte
	g byte
	b byte
}

func colorToSliderRGB(color types.TColor) sliderRGB {
	return sliderRGB{colors.Red(color), colors.Green(color), colors.Blue(color)}
}

func blendSliderPixel(dstR, dstG, dstB byte, src sliderRGB, coverage, alpha float64) (byte, byte, byte) {
	a := coverage * alpha
	if a <= 0 {
		return dstR, dstG, dstB
	}
	if a > 1 {
		a = 1
	}
	return byte(float64(dstR)*(1-a) + float64(src.r)*a + 0.5),
		byte(float64(dstG)*(1-a) + float64(src.g)*a + 0.5),
		byte(float64(dstB)*(1-a) + float64(src.b)*a + 0.5)
}

func roundedRectCoverage(px, py, scale int32, left, top, right, bottom, radius float64) float64 {
	if right <= left || bottom <= top {
		return 0
	}
	maxRadius := (right - left) / 2
	if hRadius := (bottom - top) / 2; hRadius < maxRadius {
		maxRadius = hRadius
	}
	if radius > maxRadius {
		radius = maxRadius
	}
	if radius < 0 {
		radius = 0
	}

	hit := int32(0)
	total := scale * scale
	step := 1 / float64(scale)
	for sy := int32(0); sy < scale; sy++ {
		y := float64(py) + (float64(sy)+0.5)*step
		for sx := int32(0); sx < scale; sx++ {
			x := float64(px) + (float64(sx)+0.5)*step
			if pointInRoundedRect(x, y, left, top, right, bottom, radius) {
				hit++
			}
		}
	}
	return float64(hit) / float64(total)
}

func pointInRoundedRect(x, y, left, top, right, bottom, radius float64) bool {
	if x < left || x > right || y < top || y > bottom {
		return false
	}
	cx := x
	if cx < left+radius {
		cx = left + radius
	} else if cx > right-radius {
		cx = right - radius
	}
	cy := y
	if cy < top+radius {
		cy = top + radius
	} else if cy > bottom-radius {
		cy = bottom - radius
	}
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func button(parent lcl.IWinControl, x, y, w int32, caption string, fn func()) lcl.IButton {
	b := lcl.NewButton(parent)
	b.SetParent(parent)
	b.SetBounds(x, y, w, 28)
	b.SetCaption(caption)
	b.SetOnClick(func(lcl.IObject) { fn() })
	return b
}

func parseFloat(s string, fallback float64) float64 {
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err != nil {
		return fallback
	}
	return v
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func formatTime(t float64) string {
	if t < 0 {
		t = 0
	}
	h := int(t) / 3600
	m := (int(t) % 3600) / 60
	s := int(t) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func trimForLabel(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max-1]) + "..."
}

func blank(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func screenshotName() string {
	name := "mpv-screenshot-%F-%P.png"
	if p, err := os.Getwd(); err == nil {
		return filepath.Join(p, name)
	}
	return name
}

func (m *TMPVForm) FormCloseQuery(sender lcl.IObject, canClose *bool) {
	if m.eventTimer != nil {
		m.eventTimer.SetEnabled(false)
	}
	if ctx != nil {
		for _, p := range observed {
			C.mpv_unobserve_property(ctx, C.uint64_t(p.id))
		}
		if statsClient != nil {
			C.mpv_destroy(statsClient)
			statsClient = nil
		}
		C.mpv_terminate_destroy(ctx)
		ctx = nil
	}
	*canClose = true
}
