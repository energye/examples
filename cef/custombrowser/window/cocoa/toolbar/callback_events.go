package toolbar

type callback struct {
	cb func(ctx *ToolbarCallbackContext) *GoData
}

func MakeNotifyEvent(cb NotifyEvent) *callback {
	return &callback{
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}
