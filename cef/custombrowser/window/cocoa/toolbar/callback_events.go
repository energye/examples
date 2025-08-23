package toolbar

type Callback struct {
	cb func(ctx *ToolbarCallbackContext) *GoData
}

func MakeNotifyEvent(cb NotifyEvent) *Callback {
	return &Callback{
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}
