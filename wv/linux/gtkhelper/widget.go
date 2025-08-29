package gtkhelper

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"errors"
	"unsafe"
)

// IWidget is an interface type implemented by all structs
// embedding a Widget.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkWidget.
type IWidget interface {
	toWidget() *C.GtkWidget
	ToWidget() *Widget
}

// Widget is a representation of GTK's GtkWidget.
type Widget struct {
	InitiallyUnowned
}

func wrapWidget(obj *Object) *Widget {
	if obj == nil {
		return nil
	}

	return &Widget{InitiallyUnowned{obj}}
}

// native returns a pointer to the underlying GtkWidget.
func (v *Widget) native() *C.GtkWidget {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkWidget(p)
}

func (v *Widget) toWidget() *C.GtkWidget {
	if v == nil {
		return nil
	}
	return v.native()
}

// ToWidget is a helper getter, e.g.: it returns *gtk.Label as a *gtk.Widget.
// In other cases, where you have a gtk.IWidget, use the type assertion.
func (v *Widget) ToWidget() *Widget {
	return v
}

// GetPreferredHeightForWidth is a wrapper around gtk_widget_get_preferred_height_for_width().
func (v *Widget) GetPreferredHeightForWidth(width int) (int, int) {
	var minimum, natural C.gint
	C.gtk_widget_get_preferred_height_for_width(
		v.native(),
		C.gint(width),
		&minimum,
		&natural)
	return int(minimum), int(natural)
}

// GetPreferredWidthForHeight is a wrapper around gtk_widget_get_preferred_width_for_height().
func (v *Widget) GetPreferredWidthForHeight(height int) (int, int) {

	var minimum, natural C.gint

	C.gtk_widget_get_preferred_width_for_height(
		v.native(),
		C.gint(height),
		&minimum,
		&natural)
	return int(minimum), int(natural)
}

// GetPreferredHeight is a wrapper around gtk_widget_get_preferred_height().
func (v *Widget) GetPreferredHeight() (int, int) {
	var minimum, natural C.gint
	C.gtk_widget_get_preferred_height(v.native(), &minimum, &natural)
	return int(minimum), int(natural)
}

// GetPreferredWidth is a wrapper around gtk_widget_get_preferred_width().
func (v *Widget) GetPreferredWidth() (int, int) {
	var minimum, natural C.gint
	C.gtk_widget_get_preferred_width(v.native(), &minimum, &natural)
	return int(minimum), int(natural)
}

// GetHAlign is a wrapper around gtk_widget_get_halign().
func (v *Widget) GetHAlign() Align {
	c := C.gtk_widget_get_halign(v.native())
	return Align(c)
}

// SetHAlign is a wrapper around gtk_widget_set_halign().
func (v *Widget) SetHAlign(align Align) {
	C.gtk_widget_set_halign(v.native(), C.GtkAlign(align))
}

// GetVAlign is a wrapper around gtk_widget_get_valign().
func (v *Widget) GetVAlign() Align {
	c := C.gtk_widget_get_valign(v.native())
	return Align(c)
}

// SetVAlign is a wrapper around gtk_widget_set_valign().
func (v *Widget) SetVAlign(align Align) {
	C.gtk_widget_set_valign(v.native(), C.GtkAlign(align))
}

// GetMarginTop is a wrapper around gtk_widget_get_margin_top().
func (v *Widget) GetMarginTop() int {
	c := C.gtk_widget_get_margin_top(v.native())
	return int(c)
}

// SetMarginTop is a wrapper around gtk_widget_set_margin_top().
func (v *Widget) SetMarginTop(margin int) {
	C.gtk_widget_set_margin_top(v.native(), C.gint(margin))
}

// GetMarginBottom is a wrapper around gtk_widget_get_margin_bottom().
func (v *Widget) GetMarginBottom() int {
	c := C.gtk_widget_get_margin_bottom(v.native())
	return int(c)
}

// SetMarginBottom is a wrapper around gtk_widget_set_margin_bottom().
func (v *Widget) SetMarginBottom(margin int) {
	C.gtk_widget_set_margin_bottom(v.native(), C.gint(margin))
}

// GetHExpand is a wrapper around gtk_widget_get_hexpand().
func (v *Widget) GetHExpand() bool {
	c := C.gtk_widget_get_hexpand(v.native())
	return GoBool(c)
}

// SetHExpand is a wrapper around gtk_widget_set_hexpand().
func (v *Widget) SetHExpand(expand bool) {
	C.gtk_widget_set_hexpand(v.native(), CBool(expand))
}

// TODO:
// gtk_widget_get_hexpand_set().
// gtk_widget_set_hexpand_set().

// GetVExpand is a wrapper around gtk_widget_get_vexpand().
func (v *Widget) GetVExpand() bool {
	c := C.gtk_widget_get_vexpand(v.native())
	return GoBool(c)
}

// SetVExpand is a wrapper around gtk_widget_set_vexpand().
func (v *Widget) SetVExpand(expand bool) {
	C.gtk_widget_set_vexpand(v.native(), CBool(expand))
}

// TODO:
// gtk_widget_get_vexpand_set().
// gtk_widget_set_vexpand_set().
// gtk_widget_queue_compute_expand().
// gtk_widget_compute_expand().

// GetRealized is a wrapper around gtk_widget_get_realized().
func (v *Widget) GetRealized() bool {
	c := C.gtk_widget_get_realized(v.native())
	return GoBool(c)
}

// SetRealized is a wrapper around gtk_widget_set_realized().
func (v *Widget) SetRealized(realized bool) {
	C.gtk_widget_set_realized(v.native(), CBool(realized))
}

// GetHasWindow is a wrapper around gtk_widget_get_has_window().
func (v *Widget) GetHasWindow() bool {
	c := C.gtk_widget_get_has_window(v.native())
	return GoBool(c)
}

// SetHasWindow is a wrapper around gtk_widget_set_has_window().
func (v *Widget) SetHasWindow(hasWindow bool) {
	C.gtk_widget_set_has_window(v.native(), CBool(hasWindow))
}

// ShowNow is a wrapper around gtk_widget_show_now().
func (v *Widget) ShowNow() {
	C.gtk_widget_show_now(v.native())
}

// ShowAll is a wrapper around gtk_widget_show_all().
func (v *Widget) ShowAll() {
	C.gtk_widget_show_all(v.native())
}

// SetNoShowAll is a wrapper around gtk_widget_set_no_show_all().
func (v *Widget) SetNoShowAll(noShowAll bool) {
	C.gtk_widget_set_no_show_all(v.native(), CBool(noShowAll))
}

// GetNoShowAll is a wrapper around gtk_widget_get_no_show_all().
func (v *Widget) GetNoShowAll() bool {
	c := C.gtk_widget_get_no_show_all(v.native())
	return GoBool(c)
}

// Map is a wrapper around gtk_widget_map().
func (v *Widget) Map() {
	C.gtk_widget_map(v.native())
}

// Unmap is a wrapper around gtk_widget_unmap().
func (v *Widget) Unmap() {
	C.gtk_widget_unmap(v.native())
}

// TODO:
//void gtk_widget_draw(GtkWidget *widget, cairo_t *cr);
//void gtk_widget_queue_resize(GtkWidget *widget);
//void gtk_widget_queue_resize_no_redraw(GtkWidget *widget);
// gtk_widget_queue_allocate().

// Realize is a wrapper around gtk_widget_realize().
func (v *Widget) Realize() {
	C.gtk_widget_realize(v.native())
}

// Unrealize is a wrapper around gtk_widget_unrealize().
func (v *Widget) Unrealize() {
	C.gtk_widget_unrealize(v.native())
}

// Event() is a wrapper around gtk_widget_event().
func (v *Widget) Event(event *Event) bool {
	c := C.gtk_widget_event(v.native(),
		(*C.GdkEvent)(unsafe.Pointer(event.Native())))
	return GoBool(c)
}

// Activate() is a wrapper around gtk_widget_activate().
func (v *Widget) Activate() bool {
	return GoBool(C.gtk_widget_activate(v.native()))
}

// IsFocus() is a wrapper around gtk_widget_is_focus().
func (v *Widget) IsFocus() bool {
	return GoBool(C.gtk_widget_is_focus(v.native()))
}

// GrabFocus() is a wrapper around gtk_widget_grab_focus().
func (v *Widget) GrabFocus() {
	C.gtk_widget_grab_focus(v.native())
}

// GrabDefault() is a wrapper around gtk_widget_grab_default().
func (v *Widget) GrabDefault() {
	C.gtk_widget_grab_default(v.native())
}

// SetName() is a wrapper around gtk_widget_set_name().
func (v *Widget) SetName(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_widget_set_name(v.native(), (*C.gchar)(cstr))
}

// GetName() is a wrapper around gtk_widget_get_name().  A non-nil
// error is returned in the case that gtk_widget_get_name returns NULL to
// differentiate between NULL and an empty string.
func (v *Widget) GetName() (string, error) {
	c := C.gtk_widget_get_name(v.native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// GetSensitive is a wrapper around gtk_widget_get_sensitive().
func (v *Widget) GetSensitive() bool {
	c := C.gtk_widget_get_sensitive(v.native())
	return GoBool(c)
}

// IsSensitive is a wrapper around gtk_widget_is_sensitive().
func (v *Widget) IsSensitive() bool {
	c := C.gtk_widget_is_sensitive(v.native())
	return GoBool(c)
}

// SetSensitive is a wrapper around gtk_widget_set_sensitive().
func (v *Widget) SetSensitive(sensitive bool) {
	C.gtk_widget_set_sensitive(v.native(), CBool(sensitive))
}

// GetVisible is a wrapper around gtk_widget_get_visible().
func (v *Widget) GetVisible() bool {
	c := C.gtk_widget_get_visible(v.native())
	return GoBool(c)
}

// SetVisible is a wrapper around gtk_widget_set_visible().
func (v *Widget) SetVisible(visible bool) {
	C.gtk_widget_set_visible(v.native(), CBool(visible))
}

// SetParent is a wrapper around gtk_widget_set_parent().
func (v *Widget) SetParent(parent IWidget) {
	C.gtk_widget_set_parent(v.native(), parent.toWidget())
}

// GetParent is a wrapper around gtk_widget_get_parent().
func (v *Widget) GetParent() IWidget {
	c := C.gtk_widget_get_parent(v.native())
	if c == nil {
		return nil
	}
	return wrapWidget(ToGoObject(unsafe.Pointer(c)))
}

// SetSizeRequest is a wrapper around gtk_widget_set_size_request().
func (v *Widget) SetSizeRequest(width, height int) {
	C.gtk_widget_set_size_request(v.native(), C.gint(width), C.gint(height))
}

// GetSizeRequest is a wrapper around gtk_widget_get_size_request().
func (v *Widget) GetSizeRequest() (width, height int) {
	var w, h C.gint
	C.gtk_widget_get_size_request(v.native(), &w, &h)
	return int(w), int(h)
}

// SetEvents is a wrapper around gtk_widget_set_events().
func (v *Widget) SetEvents(events int) {
	C.gtk_widget_set_events(v.native(), C.gint(events))
}

// GetEvents is a wrapper around gtk_widget_get_events().
func (v *Widget) GetEvents() int {
	return int(C.gtk_widget_get_events(v.native()))
}

// AddEvents is a wrapper around gtk_widget_add_events().
func (v *Widget) AddEvents(events int) {
	C.gtk_widget_add_events(v.native(), C.gint(events))
}

// TODO:
/*
// gtk_widget_set_device_events().
func (v *Widget) SetDeviceEvents() {
}
*/

/*
// gtk_widget_get_device_events().
func (v *Widget) GetDeviceEvents() {
}
*/

/*
// gtk_widget_add_device_events().
func (v *Widget) AddDeviceEvents() {
}
*/

// FreezeChildNotify is a wrapper around gtk_widget_freeze_child_notify().
func (v *Widget) FreezeChildNotify() {
	C.gtk_widget_freeze_child_notify(v.native())
}

// ThawChildNotify is a wrapper around gtk_widget_thaw_child_notify().
func (v *Widget) ThawChildNotify() {
	C.gtk_widget_thaw_child_notify(v.native())
}

// HasDefault is a wrapper around gtk_widget_has_default().
func (v *Widget) HasDefault() bool {
	c := C.gtk_widget_has_default(v.native())
	return GoBool(c)
}

// HasFocus is a wrapper around gtk_widget_has_focus().
func (v *Widget) HasFocus() bool {
	c := C.gtk_widget_has_focus(v.native())
	return GoBool(c)
}

// HasVisibleFocus is a wrapper around gtk_widget_has_visible_focus().
func (v *Widget) HasVisibleFocus() bool {
	c := C.gtk_widget_has_visible_focus(v.native())
	return GoBool(c)
}

// HasGrab is a wrapper around gtk_widget_has_grab().
func (v *Widget) HasGrab() bool {
	c := C.gtk_widget_has_grab(v.native())
	return GoBool(c)
}

// IsDrawable is a wrapper around gtk_widget_is_drawable().
func (v *Widget) IsDrawable() bool {
	c := C.gtk_widget_is_drawable(v.native())
	return GoBool(c)
}

// IsToplevel is a wrapper around gtk_widget_is_toplevel().
func (v *Widget) IsToplevel() bool {
	c := C.gtk_widget_is_toplevel(v.native())
	return GoBool(c)
}

// GetToplevel is a wrapper around gtk_widget_get_toplevel().
func (v *Widget) GetToplevel() IWidget {
	c := C.gtk_widget_get_toplevel(v.native())
	if c == nil {
		return nil
	}
	return wrapWidget(ToGoObject(unsafe.Pointer(c)))
}

// TODO:
// gtk_widget_get_ancestor().
// gtk_widget_get_visual().
// gtk_widget_is_ancestor().

// GetTooltipMarkup is a wrapper around gtk_widget_get_tooltip_markup().
// A non-nil error is returned in the case that gtk_widget_get_tooltip_markup
// returns NULL to differentiate between NULL and an empty string.
func (v *Widget) GetTooltipMarkup() (string, error) {
	c := C.gtk_widget_get_tooltip_markup(v.native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// SetTooltipMarkup is a wrapper around gtk_widget_set_tooltip_markup().
func (v *Widget) SetTooltipMarkup(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_widget_set_tooltip_markup(v.native(), (*C.gchar)(cstr))
}

// GetTooltipText is a wrapper around gtk_widget_get_tooltip_text().
// A non-nil error is returned in the case that
// gtk_widget_get_tooltip_text returns NULL to differentiate between NULL
// and an empty string.
func (v *Widget) GetTooltipText() (string, error) {
	c := C.gtk_widget_get_tooltip_text(v.native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// SetTooltipText is a wrapper around gtk_widget_set_tooltip_text().
func (v *Widget) SetTooltipText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_widget_set_tooltip_text(v.native(), (*C.gchar)(cstr))
}

// TranslateCoordinates is a wrapper around gtk_widget_translate_coordinates().
func (v *Widget) TranslateCoordinates(dest IWidget, srcX, srcY int) (destX, destY int, e error) {
	var cdest *C.GtkWidget = nil
	if dest != nil {
		cdest = dest.toWidget()
	}
	var cdestX, cdestY C.gint
	c := C.gtk_widget_translate_coordinates(v.native(), cdest, C.gint(srcX), C.gint(srcY), &cdestX, &cdestY)
	if !GoBool(c) {
		return 0, 0, errors.New("translate coordinates failed")
	}
	return int(cdestX), int(cdestY), nil
}

// SetAppPaintable is a wrapper around gtk_widget_set_app_paintable().
func (v *Widget) SetAppPaintable(paintable bool) {
	C.gtk_widget_set_app_paintable(v.native(), CBool(paintable))
}

// GetAppPaintable is a wrapper around gtk_widget_get_app_paintable().
func (v *Widget) GetAppPaintable() bool {
	c := C.gtk_widget_get_app_paintable(v.native())
	return GoBool(c)
}

// QueueDraw is a wrapper around gtk_widget_queue_draw().
func (v *Widget) QueueDraw() {
	C.gtk_widget_queue_draw(v.native())
}

// GetAllocation is a wrapper around gtk_widget_get_allocation().
func (v *Widget) GetAllocation() *Allocation {
	var a Allocation
	C.gtk_widget_get_allocation(v.native(), a.native())
	return &a
}

// SetAllocation is a wrapper around gtk_widget_set_allocation().
func (v *Widget) SetAllocation(allocation *Allocation) {
	C.gtk_widget_set_allocation(v.native(), allocation.native())
}

// SizeAllocate is a wrapper around gtk_widget_size_allocate().
func (v *Widget) SizeAllocate(allocation *Allocation) {
	C.gtk_widget_size_allocate(v.native(), allocation.native())
}

// TODO:
// gtk_widget_size_allocate_with_baseline().

// SetStateFlags is a wrapper around gtk_widget_set_state_flags().
func (v *Widget) SetStateFlags(stateFlags StateFlags, clear bool) {
	C.gtk_widget_set_state_flags(v.native(), C.GtkStateFlags(stateFlags), CBool(clear))
}

func (v *Widget) UnsetStateFlags(stateFlags StateFlags) {
	C.gtk_widget_unset_state_flags(v.native(), C.GtkStateFlags(stateFlags))
}

func (v *Widget) GetStateFlags() StateFlags {
	return StateFlags(C.gtk_widget_get_state_flags(v.native()))
}

/*
 * GtkAllocation
 */

// Allocation is a representation of GTK's GtkAllocation type.
type Allocation struct {
	Rectangle
}

// Native returns a pointer to the underlying GtkAllocation.
func (v *Allocation) native() *C.GtkAllocation {
	return (*C.GtkAllocation)(unsafe.Pointer(&v.GdkRectangle))
}
