package toolbar

type callback struct {
	cb func(ctx *ToolbarCallbackContext) *GoData
}

func makeWindowDidResizeAction(cb WindowDidResize) *callback {
	return &callback{
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}

func makeToolbarDefaultItemIdentifiers(cb ToolbarDefaultItemIdentifiers) *callback {
	return &callback{
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}
