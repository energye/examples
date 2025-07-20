package main

import (
	"bytes"
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"strconv"
	"time"
	"unsafe"
)

type TMainForm struct {
	lcl.TEngForm
	VST            lcl.ILazVirtualStringTree
	ClearButton    lcl.IButton
	AddOneButton   lcl.IButton
	Edit1          lcl.IEdit
	AddChildButton lcl.IButton
	Label1         lcl.ILabel
}

var (
	mainForm TMainForm
)

func init() {
	TestLoadLibPath()
}

func main() {
	lcl.Init(nil, nil)
	lcl.Application.SetScaled(true)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("minimal")
	m.WorkAreaCenter()
	m.SetHeight(500)
	m.SetWidth(380)

	labelText := "上次操作持续时间："
	m.Label1 = lcl.NewLabel(m)
	m.Label1.SetParent(m)
	m.Label1.SetLeft(8)
	m.Label1.SetTop(8)
	m.Label1.SetCaption(labelText)

	m.VST = lcl.NewLazVirtualStringTree(m)
	m.VST.SetParent(m)
	m.VST.SetHeight(365)
	m.VST.SetWidth(365)
	m.VST.SetLeft(8)
	m.VST.SetTop(27)
	m.VST.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))
	m.VST.SetAnimationDuration(200)
	m.VST.SetAutoExpandDelay(1000)
	m.VST.SetAutoScrollDelay(1000)
	m.VST.SetAutoScrollInterval(1)
	m.VST.SetEditDelay(1000)

	m.Edit1 = lcl.NewEdit(m)
	m.Edit1.SetParent(m)
	m.Edit1.SetLeft(8)
	m.Edit1.SetWidth(80)
	m.Edit1.SetTop(400)
	m.Edit1.SetCaption("1")
	m.Edit1.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))

	m.AddOneButton = lcl.NewButton(m)
	m.AddOneButton.SetParent(m)
	m.AddOneButton.SetLeft(100)
	m.AddOneButton.SetTop(400)
	m.AddOneButton.SetWidth(165)
	m.AddOneButton.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.AddOneButton.SetCaption("向根节点添加节点")

	m.AddChildButton = lcl.NewButton(m)
	m.AddChildButton.SetParent(m)
	m.AddChildButton.SetLeft(100)
	m.AddChildButton.SetTop(435)
	m.AddChildButton.SetWidth(165)
	m.AddChildButton.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.AddChildButton.SetCaption("添加节点作为子节点")

	m.ClearButton = lcl.NewButton(m)
	m.ClearButton.SetParent(m)
	m.ClearButton.SetLeft(100)
	m.ClearButton.SetTop(465)
	m.ClearButton.SetWidth(165)
	m.ClearButton.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.ClearButton.SetCaption("清除节点")

	//
	// 在 Go 里，虚拟树的数据需要存到临时集合里
	dataNodeList := make(map[types.PVirtualNode]string)
	// 始终需要为OnGetText事件设置一个处理程序，因为它为树结构提供了要显示的字符串数据。
	m.VST.SetOnGetText(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode, column int32, textType types.TVSTTextType, cellText *string) {
		dataPtr := sender.GetNodeData(node)
		if dataPtr != 0 {
			// 在集合里取出数据，并返回, node 做为每个节点 key
			*cellText = dataNodeList[node]
		}
		//fmt.Println("OnGetText")
	})
	// 构建节点标题。此事件针对每个节点触发一次，但以异步方式呈现，即节点在显示时触发，而非添加时。
	//buf := strings.Builder{}
	buf := bytes.Buffer{}
	m.VST.SetOnInitNode(func(sender lcl.IBaseVirtualTree, parentNode types.PVirtualNode, node types.PVirtualNode, initialStates *types.TVirtualNodeInitStates) {
		dataPtr := sender.GetNodeData(node)
		if dataPtr != 0 {
			nd := node.ToGo()
			// node 做为每个节点 key
			buf.WriteString("Level ")
			buf.WriteString(strconv.Itoa(int(sender.GetNodeLevel(node))))
			buf.WriteString(", Index ")
			buf.WriteString(strconv.Itoa(int(nd.Index)))
			dataNodeList[node] = buf.String()
			buf.Reset()
		}
		//fmt.Println("OnInitNode")
	})

	m.VST.SetOnFreeNode(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode) {
		dataPtr := sender.GetNodeData(node)
		if dataPtr != 0 {
			data := *(*tMyRec)(unsafe.Pointer(dataPtr))
			data.Caption = 0
		}
	})

	// 注意: 放在事件之后执行
	// 让树知道我们需要多少数据空间。
	m.VST.SetNodeDataSize(int32(unsafe.Sizeof(tMyRec{})))
	// 设定初始节点数量。
	m.VST.SetRootNodeCount(20)

	// 按钮事件
	m.ClearButton.SetOnClick(func(sender lcl.IObject) {
		lcl.Screen.SetCursor(types.CrHourGlass)
		start := time.Now()
		m.VST.Clear()
		dataNodeList = make(map[types.PVirtualNode]string)
		m.Label1.SetCaption(fmt.Sprintf("%v%v ms %v ns", labelText, time.Now().Sub(start).Nanoseconds()/1000000, time.Now().Sub(start).Nanoseconds()))
		lcl.Screen.SetCursor(types.CrDefault)
	})

	m.AddOneButton.SetOnClick(func(sender lcl.IObject) {
		lcl.Screen.SetCursor(types.CrHourGlass)
		start := time.Now()
		count, _ := strconv.Atoi(m.Edit1.Text())
		m.VST.SetRootNodeCount(m.VST.RootNodeCount() + uint32(count))
		m.Label1.SetCaption(fmt.Sprintf("%v%v ms %v ns", labelText, time.Now().Sub(start).Nanoseconds()/1000000, time.Now().Sub(start).Nanoseconds()))
		lcl.Screen.SetCursor(types.CrDefault)
	})

	m.AddChildButton.SetOnClick(func(sender lcl.IObject) {
		lcl.Screen.SetCursor(types.CrHourGlass)
		start := time.Now()
		if nodePtr := m.VST.FocusedNode(); nodePtr != 0 {
			count, _ := strconv.Atoi(m.Edit1.Text())
			m.VST.SetChildCount(nodePtr, m.VST.ChildCount(nodePtr)+uint32(count))
			m.VST.SetExpanded(nodePtr, true)
			m.VST.InvalidateToBottom(nodePtr)
		}
		m.Label1.SetCaption(fmt.Sprintf("%v%v ms %v ns", labelText, time.Now().Sub(start).Nanoseconds()/1000000, time.Now().Sub(start).Nanoseconds()))
		lcl.Screen.SetCursor(types.CrDefault)
	})
}

type TMyRec struct {
	Caption string
}

// tMyRec = ^TMyRec
type tMyRec struct {
	Caption uintptr // string
}
