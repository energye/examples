package toolbar

type Callback struct {
	type_ TccType
	cb    func(ctx *ToolbarCallbackContext) *GoArguments
}

func MakeNotifyEvent(cb NotifyEvent) *Callback {
	return &Callback{
		cb: func(ctx *ToolbarCallbackContext) *GoArguments {
			return cb(ctx.Identifier, ctx.Owner, ctx.Sender)
		},
	}
}

func MakeTextChangeEvent(cb TextEvent) *Callback {
	return &Callback{
		type_: TCCTextDidChange,
		cb: func(ctx *ToolbarCallbackContext) *GoArguments {
			return cb(ctx.Identifier, ctx.Value, ctx.Owner, ctx.Sender)
		},
	}
}

func MakeTextCommitEvent(cb TextEvent) *Callback {
	return &Callback{
		type_: TCCTextDidEndEditing,
		cb: func(ctx *ToolbarCallbackContext) *GoArguments {
			return cb(ctx.Identifier, ctx.Value, ctx.Owner, ctx.Sender)
		},
	}
}

//func MakeDelegateToolbarEvent(cb DelegateToolbarEvent) *Callback {
//	return &Callback{
//		type_: TCCTextDidEndEditing,
//		cb: func(ctx *ToolbarCallbackContext) *GoArguments {
//			return cb(ctx.Identifier, ctx.Value, ctx.Owner, ctx.Sender)
//		},
//	}
//}
