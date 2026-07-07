package ui

import (
	"strings"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
)

func configureOpenGLControl(ctrl lcl.IOpenGLControl) {
	if ctrl == nil {
		return
	}
	ctrl.SetRGBA(true)
	if strings.EqualFold(libname.UseWS, "gtk3") {
		ctrl.SetOpenGLMajorVersion(3)
		ctrl.SetOpenGLMinorVersion(2)
	} else {
		ctrl.SetOpenGLMajorVersion(2)
		ctrl.SetOpenGLMinorVersion(1)
	}
	ctrl.SetAlphaBits(8)
	ctrl.SetDepthBits(0)
	ctrl.SetStencilBits(8)
	ctrl.SetMultiSampling(0)
	ctrl.SetAutoResizeViewport(false)
}
