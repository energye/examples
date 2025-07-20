package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"unsafe"
)

type TMainForm struct {
	lcl.TEngForm
	VST                lcl.ILazVirtualStringTree
	ListBox            lcl.IListBox
	ShowHeaderCheckBox lcl.ICheckBox
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
	m.SetCaption("dragdrop")
	m.WorkAreaCenter()
	m.SetHeight(350)
	m.SetWidth(500)

	m.ShowHeaderCheckBox = lcl.NewCheckBox(m)
	m.ShowHeaderCheckBox.SetParent(m)
	m.ShowHeaderCheckBox.SetLeft(8)
	m.ShowHeaderCheckBox.SetTop(5)
	m.ShowHeaderCheckBox.SetCaption("显示头部")
	m.ShowHeaderCheckBox.SetOnChange(func(sender lcl.IObject) {
		if m.ShowHeaderCheckBox.Checked() {
			m.SetFormStyle(types.FsSystemStayOnTop)
		} else {
			m.SetFormStyle(types.FsNormal)
		}
	})

	m.VST = lcl.NewLazVirtualStringTree(m)
	m.VST.SetParent(m)
	m.VST.SetHeight(315)
	m.VST.SetWidth(240)
	m.VST.SetLeft(8)
	m.VST.SetTop(27)
	m.VST.SetDragMode(types.DmAutomatic)
	m.VST.SetDragType(types.DtVCL)
	m.VST.SetDragOperations(types.NewSet(types.DoCopy, types.DoMove))
	m.VST.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkBottom))

	m.ListBox = lcl.NewListBox(m)
	m.ListBox.SetParent(m)
	m.ListBox.SetHeight(315)
	m.ListBox.SetWidth(240)
	m.ListBox.SetLeft(250)
	m.ListBox.SetTop(27)
	m.ListBox.SetDragMode(types.DmAutomatic)
	m.ListBox.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkBottom))
	items := lcl.NewStringList()
	defer items.Free()
	for i := 0; i < 10; i++ {
		items.Add(fmt.Sprintf("List Item %v", i))
	}
	m.ListBox.SetItems(items)

	dataNodeList := make(map[types.PVirtualNode]string)

	m.ListBox.SetOnDragOver(func(sender lcl.IObject, source lcl.IObject, X int32, Y int32, state types.TDragState, accept *bool) {
		*accept = source.Equals(m.VST) || source.Equals(m.ListBox)
		//fmt.Println("ListBox.OnDragOver", *accept)
	})
	m.ListBox.SetOnDragDrop(func(sender lcl.IObject, source lcl.IObject, X int32, Y int32) {
		//fmt.Println("ListBox.OnDragDrop")
		if source.Equals(m.VST) {
			nodePtr := m.VST.FocusedNode()
			if nodePtr != 0 {
				m.ListBox.Items().Append(m.VST.Text(nodePtr, 0))
			}
		}
	})

	m.VST.SetOnGetText(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode, column int32, textType types.TVSTTextType, cellText *string) {
		dataPtr := sender.GetNodeData(node)
		if dataPtr != 0 {
			// 在集合里取出数据，并返回, node 做为每个节点 key
			*cellText = dataNodeList[node]
		}
	})
	m.VST.SetOnInitNode(func(sender lcl.IBaseVirtualTree, parentNode types.PVirtualNode, node types.PVirtualNode, initialStates *types.TVirtualNodeInitStates) {
		dataPtr := sender.GetNodeData(node)
		if dataPtr != 0 {
			nd := node.ToGo()
			// node 做为每个节点 key
			dataNodeList[node] = fmt.Sprintf("Level %v, Index %v", sender.GetNodeLevel(node), nd.Index)
		}
	})
	m.VST.SetOnGetNodeDataSize(func(sender lcl.IBaseVirtualTree, nodeDataSize *int32) {
		*nodeDataSize = int32(unsafe.Sizeof(tNodeData{}))
	})
	m.VST.SetOnDragDrop(func(sender lcl.IBaseVirtualTree, source lcl.IObject, dataObject lcl.IDataObject, formats lcl.IFormatArray, shift types.TShiftState, pt types.TPoint, effect *uint32, mode types.TDropMode) {
		//fmt.Println("VST.OnDragDrop")
		var (
			nodePtr   types.PVirtualNode
			nodeTitle string
		)
		switch mode {
		case types.DmAbove:
			nodePtr = sender.InsertNode(sender.DropTargetNode(), types.AmInsertBefore, 0)
		case types.DmBelow:
			nodePtr = sender.InsertNode(sender.DropTargetNode(), types.AmInsertAfter, 0)
		case types.DmNowhere:
			return
		default:
			nodePtr = sender.AddChild(sender.DropTargetNode(), 0)
		}

		sender.ValidateNode(nodePtr, true)

		if source.Equals(m.ListBox) {
			if m.ListBox.ItemIndex() == -1 {
				nodeTitle = "Unknow Item from List"
			} else {
				nodeTitle = m.ListBox.Items().Strings(m.ListBox.ItemIndex())
			}
		} else if source.Equals(sender) {
			if sender.FocusedNode() != 0 {
				nodeTitle = m.VST.Text(sender.FocusedNode(), 0)
			} else {
				nodeTitle = "Unknow Source Node"
			}
		} else {
			nodeTitle = "Unknow Source Control"
		}
		dataNodeList[nodePtr] = nodeTitle
	})
	m.VST.SetOnDragOver(func(sender lcl.IBaseVirtualTree, source lcl.IObject, shift types.TShiftState, state types.TDragState, pt types.TPoint, mode types.TDropMode, effect *uint32, accept *bool) {
		if sender != nil {
			*accept = sender.Equals(m.VST)
		} else if source != nil {
			*accept = source.Equals(m.ListBox)
		}
		//fmt.Println("VST.OnDragOver", *accept)
	})
	// 设定初始节点数量。
	m.VST.SetRootNodeCount(20)
}

type TNodeData struct {
	Caption string
}

// tNodeData = ^TNodeData
type tNodeData struct {
	Caption uintptr // string
}
