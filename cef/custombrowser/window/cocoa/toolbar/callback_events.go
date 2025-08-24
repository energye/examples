package toolbar

type Callback struct {
	type_ TccType
	cb    func(ctx *ToolbarCallbackContext) *GoData
}

func MakeNotifyEvent(cb NotifyEvent) *Callback {
	return &Callback{
		type_: TCCClicked,
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}

func MakeTextChangeEventEvent(cb TextEvent) *Callback {
	return &Callback{
		type_: TCCTextDidChange,
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Value, ctx.Owner, ctx.Sender)
		},
	}
}
func MakeTextCommitEventEvent(cb TextEvent) *Callback {
	return &Callback{
		type_: TCCTextDidEndEditing,
		cb: func(ctx *ToolbarCallbackContext) *GoData {
			return cb(ctx.Identifier, ctx.Value, ctx.Owner, ctx.Sender)
		},
	}
}
