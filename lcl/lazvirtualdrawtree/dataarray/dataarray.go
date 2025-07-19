package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"unsafe"
)

type TMainForm struct {
	lcl.TEngForm
	VST               lcl.ILazVirtualStringTree
	AddNodeBtn        lcl.IButton
	DeleteSelectedBtn lcl.IButton
	CleanAllBtn       lcl.IButton
	ClickNodeLabel    lcl.ILabel
	ClickNodeEdit     lcl.IEdit
	FindFilterLabel   lcl.ILabel
	FindFilterEdit    lcl.IEdit
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
	m.SetCaption("dataarray")
	m.WorkAreaCenter()
	m.SetHeight(400)
	m.SetWidth(650)

	// 创建控件
	m.VST = lcl.NewLazVirtualStringTree(m)
	m.VST.SetParent(m)
	m.VST.SetHeight(300)
	m.VST.SetWidth(640)
	m.VST.SetTop(5)
	m.VST.SetLeft(5)
	m.VST.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))

	m.AddNodeBtn = lcl.NewButton(m)
	m.AddNodeBtn.SetParent(m)
	m.AddNodeBtn.SetTop(m.VST.Height() + m.VST.Top() + 5)
	m.AddNodeBtn.SetLeft(5)
	m.AddNodeBtn.SetWidth(100)
	m.AddNodeBtn.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.AddNodeBtn.SetCaption("添加 10W 节点")

	m.DeleteSelectedBtn = lcl.NewButton(m)
	m.DeleteSelectedBtn.SetParent(m)
	m.DeleteSelectedBtn.SetTop(m.VST.Height() + m.VST.Top() + 5)
	m.DeleteSelectedBtn.SetLeft(105)
	m.DeleteSelectedBtn.SetWidth(100)
	m.DeleteSelectedBtn.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.DeleteSelectedBtn.SetCaption("删除选中节点")

	m.CleanAllBtn = lcl.NewButton(m)
	m.CleanAllBtn.SetParent(m)
	m.CleanAllBtn.SetTop(m.VST.Height() + m.VST.Top() + 50)
	m.CleanAllBtn.SetLeft(5)
	m.CleanAllBtn.SetWidth(100)
	m.CleanAllBtn.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.CleanAllBtn.SetCaption("清除所有")

	m.ClickNodeLabel = lcl.NewLabel(m)
	m.ClickNodeLabel.SetParent(m)
	m.ClickNodeLabel.SetTop(m.VST.Height() + m.VST.Top() + 10)
	m.ClickNodeLabel.SetLeft(210)
	m.ClickNodeLabel.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.ClickNodeLabel.SetCaption("点击节点的数组数据：")

	m.ClickNodeEdit = lcl.NewEdit(m)
	m.ClickNodeEdit.SetParent(m)
	m.ClickNodeEdit.SetTop(m.VST.Height() + m.VST.Top() + 5)
	m.ClickNodeEdit.SetLeft(325)
	m.ClickNodeEdit.SetWidth(200)
	m.ClickNodeEdit.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))

	m.FindFilterLabel = lcl.NewLabel(m)
	m.FindFilterLabel.SetParent(m)
	m.FindFilterLabel.SetTop(m.VST.Height() + m.VST.Top() + 55)
	m.FindFilterLabel.SetLeft(210)
	m.FindFilterLabel.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))
	m.FindFilterLabel.SetCaption("按特定数组索引查找并显示节点\n键入索引以在屏幕上获取相关树节点：")

	m.FindFilterEdit = lcl.NewEdit(m)
	m.FindFilterEdit.SetParent(m)
	m.FindFilterEdit.SetTop(m.VST.Height() + m.VST.Top() + 50)
	m.FindFilterEdit.SetLeft(410)
	m.FindFilterEdit.SetWidth(115)
	m.FindFilterEdit.SetAnchors(types.NewSet(types.AkLeft, types.AkBottom))

	// 如果要手动设置必要的事件或参数，
	// 没有对象检查器。这将有助于以防万一您不小心删除了组件而且你没有时间使用对象检查器并重新排列事件或其他属性
	// 首先，遵循您可以在此处或使用对象检查器设置的属性，使其更适合标准使用
	// 显示标题列
	header := m.VST.Header()
	header.SetOptions(header.Options().Include(types.HoVisible))
	// 允许多选节点
	treeOptions := m.VST.TreeOptions()
	treeOptions.SetSelectionOptions(treeOptions.SelectionOptions().Include(types.ToMultiSelect))
	// 允许在屏幕之外进行自动多选
	treeOptions.SetAutoOptions(treeOptions.AutoOptions().Include(types.ToAutoScroll))
	// 如果在自动多选过程中1000毫秒的延迟太慢
	m.VST.SetAutoScrollDelay(100)
	// 禁用拖放操作期间自动删除移动的数据
	treeOptions.SetAutoOptions(treeOptions.AutoOptions().Exclude(types.ToAutoDeleteMovedNodes))
	// 在VST上显示背景图像
	treeOptions.SetPaintOptions(treeOptions.PaintOptions().Include(types.ToShowBackground))
	// 如果你不想显示树线
	// VST.TreeOptions.PaintOptions := VST.TreeOptions.PaintOptions -[toShowTreeLines];
	// 如果你不想显示主节点的左边距
	// VST.TreeOptions.PaintOptions := VST.TreeOptions.PaintOptions -[toShowRoot];

	// 如果要手动添加列
	columns := header.Columns()
	columns.Clear()
	type ColumnParam struct {
		Name      string
		Len       int32
		Alignment types.TAlignment
	}
	columnParams := []ColumnParam{
		{Name: "文本", Len: 150, Alignment: types.TaLeftJustify},
		{Name: "指针", Len: 300, Alignment: types.TaLeftJustify},
		{Name: "随机", Len: 120, Alignment: types.TaLeftJustify},
	}
	for i := 0; i < len(columnParams); i++ {
		col := columnParams[i]
		newCol := columns.AddToVirtualTreeColumn()
		newCol.SetText(col.Name)
		newCol.SetWidth(col.Len)
		newCol.SetAlignment(col.Alignment)
	}
	//如果你想让第二列在点击时不响应
	column1 := columns.ItemsWithColumnIndexToVirtualTreeColumn(1)
	column1.SetOptions(column1.Options().Exclude(types.CoAllowClick))
	// 注册事件
	m.VST.SetOnGetText(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode, column int32, textType types.TVSTTextType, cellText *string) {

	})
	m.VST.SetOnPaintText(func(sender lcl.IBaseVirtualTree, targetCanvas lcl.ICanvas, node types.PVirtualNode, column int32, textType types.TVSTTextType) {

	})
	m.VST.SetOnHeaderClick(func(sender lcl.IVTHeader, hitInfo lcl.TVTHeaderHitInfo) {

	})
	m.VST.SetOnFocusChanged(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode, column int32) {

	})
	m.VST.SetOnCompareNodes(func(sender lcl.IBaseVirtualTree, node1 types.PVirtualNode, node2 types.PVirtualNode, column int32, result *int32) {

	})
	m.VST.SetOnBeforeCellPaint(func(sender lcl.IBaseVirtualTree, targetCanvas lcl.ICanvas, node types.PVirtualNode, column int32, cellPaintMode types.TVTCellPaintMode,
		cellRect types.TRect, contentRect *types.TRect) {

	})
	m.VST.SetOnFreeNode(func(sender lcl.IBaseVirtualTree, node types.PVirtualNode) {

	})
	// 显示标题
	header.SetOptions(header.Options().Include(types.HoVisible))
	// 显示方向标志
	header.SetOptions(header.Options().Include(types.HoShowSortGlyphs))

	// VST 最后
	// 初始化 VST 中节点的大小这是在使用VST之前最重要的一步， 因为这是VST为节点分配所需空间的唯一方法
	m.VST.SetNodeDataSize(int32(unsafe.Sizeof(TTreeData{})))
	// 当您自己添加数据时，请确保树中没有节点
	m.VST.SetRootNodeCount(0)

	// 按钮事件
	m.AddNodeBtn.SetOnClick(func(sender lcl.IObject) {
		// 添加100000条新记录和相应的VST节点
		m.VST.BeginUpdate()

		m.VST.EndUpdate()
	})
}

func (m *TMainForm) FormAfterCreate(sender lcl.IObject) {
}

type TTreeData struct {
	DataIndex int32
}
