package domvisitor

import (
	"fmt"
	"github.com/energye/cef/cef"
)

var dv cef.IEngDomVisitor

func DomVisitor() {
	if dv == nil {
		dv = cef.NewEngDomVisitor()
		dv.SetOnDomVisitorVisit(func(document cef.ICefDomDocument) {
			fmt.Println("title:", document.GetTitle())
			body := document.GetBody()
			fmt.Println("body-InnerText:", body.GetElementInnerText())
			fmt.Println("GetNodeType:", body.GetType())
		})
	}
	v8ctx := cef.V8ContextRef.Current()
	ctxFrame := v8ctx.GetFrame()
	defer func() {
		ctxFrame.Free()
		v8ctx.Free()
	}()
	visitor := cef.AsEngDomVisitor(dv.AsIntfDomVisitor())
	ctxFrame.VisitDom(visitor)
	fmt.Println("DomVisitor end RefCount:", visitor.RefCount())
}
