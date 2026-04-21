//go:build !cgo

package src

import (
	"fmt"
	"github.com/energye/energy/v3/platform/darwin/cocoa/nocgo/notification"
	. "github.com/energye/energy/v3/platform/notification/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/types"
	"time"
)

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetWidth(800)
	m.SetHeight(650)
	m.WorkAreaCenter()
	m.SetCaption("macOS 通知功能完整示例")
	fmt.Printf("OSVersion: %+v\n", version.OSVersion)
	// 初始化通知服务
	m.notifService = notification.New()

	// 注册通知响应回调
	if notify, ok := m.notifService.(INotificationDarwin); ok {
		notify.SetOnNotificationResponse(func(result Result) {
			if result.Error != nil {
				m.appendLog(fmt.Sprintf("❌ 错误: %v\n", result.Error))
				return
			}

			resp := result.Response
			m.appendLog(fmt.Sprintf("📨 收到通知响应:\n"))
			m.appendLog(fmt.Sprintf("   ID: %s\n", resp.ID))
			m.appendLog(fmt.Sprintf("   操作: %s\n", resp.ActionIdentifier))
			m.appendLog(fmt.Sprintf("   标题: %s\n", resp.Title))
			m.appendLog(fmt.Sprintf("   副标题: %s\n", resp.Subtitle))
			m.appendLog(fmt.Sprintf("   内容: %s\n", resp.Body))

			if resp.UserText != "" {
				m.appendLog(fmt.Sprintf("   ✍️ 用户输入: %s\n", resp.UserText))
			}

			if len(resp.UserInfo) > 0 {
				m.appendLog(fmt.Sprintf("   📊 附加数据: %v\n", resp.UserInfo))
			}

			m.appendLog("\n")
		})
	}

	// 主面板
	mainPanel := lcl.NewPanel(m)
	mainPanel.SetParent(m)
	mainPanel.SetAlign(types.AlTop)
	mainPanel.SetHeight(280)

	// 标题标签
	titleLabel := lcl.NewLabel(mainPanel)
	titleLabel.SetParent(mainPanel)
	titleLabel.SetCaption("🔔 macOS 通知功能完整示例")
	titleLabel.SetLeft(240)
	titleLabel.SetTop(10)

	// 状态标签
	m.statusLabel = lcl.NewLabel(mainPanel)
	m.statusLabel.SetParent(mainPanel)
	m.statusLabel.SetCaption("就绪 - 请先请求通知权限")
	m.statusLabel.SetLeft(20)
	m.statusLabel.SetTop(50)
	m.statusLabel.SetWidth(760)
	m.statusLabel.SetTransparent(true)

	// ========== 第一行按钮 ==========
	row1Top := int32(80)

	// 按钮1: 请求权限
	btnAuth := lcl.NewButton(mainPanel)
	btnAuth.SetParent(mainPanel)
	btnAuth.SetCaption("① 请求通知权限")
	btnAuth.SetLeft(20)
	btnAuth.SetTop(row1Top)
	btnAuth.SetWidth(180)
	btnAuth.SetHint("首次使用必须点击此按钮获取权限")
	btnAuth.SetOnClick(func(sender lcl.IObject) {
		m.setStatus("正在请求权限...")
		go func() {
			authorized, err := m.notifService.RequestNotificationAuthorization()
			lcl.RunOnMainThreadAsync(func(id uint32) {
				if err != nil {
					m.setStatus(fmt.Sprintf("❌ 权限请求失败: %v", err))
					m.appendLog(fmt.Sprintf("权限请求失败: %v\n", err))
					return
				}

				if authorized {
					m.setStatus("✅ 已获得通知权限！可以发送通知了")
					m.appendLog("✅ 通知权限已授予\n")
				} else {
					m.setStatus("❌ 用户拒绝了权限，请在系统偏好设置中手动启用")
					m.appendLog("❌ 通知权限被拒绝\n")
				}
			})
		}()
	})

	// 按钮2: 检查权限
	btnCheck := lcl.NewButton(mainPanel)
	btnCheck.SetParent(mainPanel)
	btnCheck.SetCaption("② 检查权限状态")
	btnCheck.SetLeft(220)
	btnCheck.SetTop(row1Top)
	btnCheck.SetWidth(180)
	btnCheck.SetHint("查看当前通知权限状态")
	btnCheck.SetOnClick(func(sender lcl.IObject) {
		go func() {
			authorized, err := m.notifService.CheckNotificationAuthorization()
			lcl.RunOnMainThreadAsync(func(id uint32) {
				if err != nil {
					m.setStatus(fmt.Sprintf("检查失败: %v", err))
					return
				}

				if authorized {
					m.setStatus("✅ 已授权 - 可以发送通知")
					m.appendLog("权限状态: 已授权\n")
				} else {
					m.setStatus("⚠️ 未授权或已拒绝")
					m.appendLog("权限状态: 未授权\n")
				}
			})
		}()
	})

	// 按钮3: 简单通知
	btnSimple := lcl.NewButton(mainPanel)
	btnSimple.SetParent(mainPanel)
	btnSimple.SetCaption("③ 简单通知")
	btnSimple.SetLeft(420)
	btnSimple.SetTop(row1Top)
	btnSimple.SetWidth(160)
	btnSimple.SetHint("最基本的通知，无交互")
	btnSimple.SetOnClick(func(sender lcl.IObject) {
		opts := Options{
			ID:       fmt.Sprintf("simple-%d", time.Now().Unix()),
			Title:    "Hello Energy!",
			Subtitle: "简单通知示例",
			Body:     "这是一个基本的通知，没有任何交互按钮 🎉",
		}

		err := m.notifService.SendNotification(opts)
		if err != nil {
			m.setStatus(fmt.Sprintf("发送失败: %v", err))
		} else {
			m.setStatus("✅ 简单通知已发送")
			m.appendLog("📤 发送简单通知\n")
		}
	})

	// ========== 第二行按钮 ==========
	row2Top := int32(130)

	// 按钮4: 带数据的通知
	btnWithData := lcl.NewButton(mainPanel)
	btnWithData.SetParent(mainPanel)
	btnWithData.SetCaption("④ 带数据的通知")
	btnWithData.SetLeft(20)
	btnWithData.SetTop(row2Top)
	btnWithData.SetWidth(180)
	btnWithData.SetHint("通知携带自定义数据")
	btnWithData.SetOnClick(func(sender lcl.IObject) {
		opts := Options{
			ID:       fmt.Sprintf("data-%d", time.Now().Unix()),
			Title:    "任务完成",
			Subtitle: "数据处理",
			Body:     "所有文件已成功处理完毕",
			Data: map[string]interface{}{
				"type":      "task_complete",
				"fileCount": 42,
				"duration":  "2m 35s",
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			},
		}

		err := m.notifService.SendNotification(opts)
		if err != nil {
			m.setStatus(fmt.Sprintf("发送失败: %v", err))
		} else {
			m.setStatus("✅ 带数据的通知已发送")
			m.appendLog("📤 发送带数据的通知\n")
		}
	})

	// 按钮5: 成功通知
	btnSuccess := lcl.NewButton(mainPanel)
	btnSuccess.SetParent(mainPanel)
	btnSuccess.SetCaption("⑤ ✅ 成功通知")
	btnSuccess.SetLeft(220)
	btnSuccess.SetTop(row2Top)
	btnSuccess.SetWidth(160)
	btnSuccess.SetHint("绿色样式的成功提示")
	btnSuccess.SetOnClick(func(sender lcl.IObject) {
		opts := Options{
			ID:       fmt.Sprintf("success-%d", time.Now().Unix()),
			Title:    "✅ 操作成功",
			Subtitle: "保存完成",
			Body:     "您的更改已成功保存到数据库",
		}

		m.notifService.SendNotification(opts)
		m.setStatus("✅ 成功通知已发送")
		m.appendLog("📤 发送成功通知\n")
	})

	// 按钮6: 警告通知
	btnWarning := lcl.NewButton(mainPanel)
	btnWarning.SetParent(mainPanel)
	btnWarning.SetCaption("⑥ ⚠️ 警告通知")
	btnWarning.SetLeft(400)
	btnWarning.SetTop(row2Top)
	btnWarning.SetWidth(160)
	btnWarning.SetHint("黄色样式的警告提示")
	btnWarning.SetOnClick(func(sender lcl.IObject) {
		opts := Options{
			ID:       fmt.Sprintf("warning-%d", time.Now().Unix()),
			Title:    "⚠️ 存储空间不足",
			Subtitle: "系统警告",
			Body:     "磁盘空间仅剩 5%，请及时清理",
		}

		m.notifService.SendNotification(opts)
		m.setStatus("⚠️ 警告通知已发送")
		m.appendLog("📤 发送警告通知\n")
	})

	// 按钮7: 错误通知
	btnError := lcl.NewButton(mainPanel)
	btnError.SetParent(mainPanel)
	btnError.SetCaption("⑦ ❌ 错误通知")
	btnError.SetLeft(580)
	btnError.SetTop(row2Top)
	btnError.SetWidth(160)
	btnError.SetHint("红色样式的错误提示")
	btnError.SetOnClick(func(sender lcl.IObject) {
		opts := Options{
			ID:       fmt.Sprintf("error-%d", time.Now().Unix()),
			Title:    "❌ 连接失败",
			Subtitle: "网络错误",
			Body:     "无法连接到服务器，请检查网络设置",
		}

		m.notifService.SendNotification(opts)
		m.setStatus("❌ 错误通知已发送")
		m.appendLog("📤 发送错误通知\n")
	})

	// ========== 第三行按钮 - 交互通知 ==========
	row3Top := int32(180)

	// 按钮8: 带两个按钮的通知
	btnTwoActions := lcl.NewButton(mainPanel)
	btnTwoActions.SetParent(mainPanel)
	btnTwoActions.SetCaption("⑧ 双按钮通知")
	btnTwoActions.SetLeft(20)
	btnTwoActions.SetTop(row3Top)
	btnTwoActions.SetWidth(180)
	btnTwoActions.SetHint("包含打开和删除两个操作按钮")
	btnTwoActions.SetOnClick(func(sender lcl.IObject) {
		// 注册类别
		category := Category{
			ID: "two_button_category",
			Actions: []Action{
				{
					ID:    "open_action",
					Title: "📂 打开",
				},
				{
					ID:          "delete_action",
					Title:       "🗑️ 删除",
					Destructive: true,
				},
			},
		}

		err := m.notifService.RegisterNotificationCategory(category)
		if err != nil {
			m.setStatus(fmt.Sprintf("注册类别失败: %v", err))
			return
		}

		// 发送通知
		opts := Options{
			ID:         fmt.Sprintf("twobtn-%d", time.Now().Unix()),
			Title:      "新文档已下载",
			Subtitle:   "下载完成",
			Body:       "report.pdf 已保存到下载文件夹",
			CategoryID: "two_button_category",
			Data: map[string]interface{}{
				"filename": "report.pdf",
				"path":     "~/Downloads/report.pdf",
			},
		}

		err = m.notifService.SendNotificationWithActions(opts)
		if err != nil {
			m.setStatus(fmt.Sprintf("发送失败: %v", err))
		} else {
			m.setStatus("✅ 双按钮通知已发送（展开通知查看按钮）")
			m.appendLog("📤 发送双按钮通知\n")
		}
	})

	// 按钮9: 三个按钮的通知
	btnThreeActions := lcl.NewButton(mainPanel)
	btnThreeActions.SetParent(mainPanel)
	btnThreeActions.SetCaption("⑨ 三按钮通知")
	btnThreeActions.SetLeft(220)
	btnThreeActions.SetTop(row3Top)
	btnThreeActions.SetWidth(180)
	btnThreeActions.SetHint("包含三个操作选项")
	btnThreeActions.SetOnClick(func(sender lcl.IObject) {
		category := Category{
			ID: "three_button_category",
			Actions: []Action{
				{
					ID:    "accept_action",
					Title: "✅ 接受",
				},
				{
					ID:    "reject_action",
					Title: "❌ 拒绝",
				},
				{
					ID:          "block_action",
					Title:       "🚫 屏蔽",
					Destructive: true,
				},
			},
		}

		m.notifService.RegisterNotificationCategory(category)

		opts := Options{
			ID:         fmt.Sprintf("threebtn-%d", time.Now().Unix()),
			Title:      "好友请求",
			Subtitle:   "社交通知",
			Body:       "张三想要添加你为好友",
			CategoryID: "three_button_category",
		}

		m.notifService.SendNotificationWithActions(opts)
		m.setStatus("✅ 三按钮通知已发送")
		m.appendLog("📤 发送三按钮通知\n")
	})

	// 按钮10: 文本输入通知
	btnTextInput := lcl.NewButton(mainPanel)
	btnTextInput.SetParent(mainPanel)
	btnTextInput.SetCaption("⑩ 💬 文本输入通知")
	btnTextInput.SetLeft(420)
	btnTextInput.SetTop(row3Top)
	btnTextInput.SetWidth(180)
	btnTextInput.SetHint("可以在通知中直接输入文字回复")
	btnTextInput.SetOnClick(func(sender lcl.IObject) {
		category := Category{
			ID:               "text_input_category",
			Actions:          []Action{}, // ✅ 空数组，不添加普通按钮
			HasReplyField:    true,
			ReplyPlaceholder: "输入回复内容...",
			ReplyButtonTitle: "发送",
		}

		m.notifService.RegisterNotificationCategory(category)

		opts := Options{
			ID:         fmt.Sprintf("input-%d", time.Now().Unix()),
			Title:      "💬 新消息",
			Subtitle:   "来自李四",
			Body:       "你好，在吗？有个问题想请教",
			CategoryID: "text_input_category",
			Data: map[string]interface{}{
				"sender":   "李四",
				"chatId":   "chat_12345",
				"priority": "high",
			},
		}

		m.notifService.SendNotificationWithActions(opts)
		m.setStatus("✅ 文本输入通知已发送（展开可输入回复）")
		m.appendLog("📤 发送文本输入通知\n")
	})

	// 按钮11: 快捷回复 + 输入框
	btnQuickReply := lcl.NewButton(mainPanel)
	btnQuickReply.SetParent(mainPanel)
	btnQuickReply.SetCaption("⑪ ⚡ 快捷回复+输入")
	btnQuickReply.SetLeft(620)
	btnQuickReply.SetTop(row3Top)
	btnQuickReply.SetWidth(160)
	btnQuickReply.SetHint("快捷按钮 + 自定义输入")
	btnQuickReply.SetOnClick(func(sender lcl.IObject) {
		category := Category{
			ID: "quick_reply_category",
			Actions: []Action{
				{
					ID:    "quick_yes",
					Title: "好的👍",
				},
				{
					ID:    "quick_later",
					Title: "稍等⏰",
				},
				{
					ID:    "custom_reply",
					Title: "自定义",
				},
			},
			HasReplyField:    true,
			ReplyPlaceholder: "输入自定义回复...",
			ReplyButtonTitle: "发送",
		}

		m.notifService.RegisterNotificationCategory(category)

		opts := Options{
			ID:         fmt.Sprintf("quick-%d", time.Now().Unix()),
			Title:      "会议提醒",
			Subtitle:   "15分钟后",
			Body:       "产品评审会议即将开始",
			CategoryID: "quick_reply_category",
		}

		m.notifService.SendNotificationWithActions(opts)
		m.setStatus("✅ 快捷回复通知已发送")
		m.appendLog("📤 发送快捷回复+输入通知\n")
	})

	// ========== 第四行按钮 - 管理功能 ==========
	row4Top := int32(230)

	// 按钮12: 清除所有通知
	btnClear := lcl.NewButton(mainPanel)
	btnClear.SetParent(mainPanel)
	btnClear.SetCaption("🗑️ 清除所有通知")
	btnClear.SetLeft(20)
	btnClear.SetTop(row4Top)
	btnClear.SetWidth(180)
	btnClear.SetHint("移除所有已显示和待显示的通知")
	btnClear.SetOnClick(func(sender lcl.IObject) {
		m.notifService.RemoveAllDeliveredNotifications()
		m.notifService.RemoveAllPendingNotifications()
		m.setStatus("🗑️ 已清除所有通知")
		m.appendLog("🗑️ 清除所有通知\n")
	})

	// 按钮13: 批量发送测试
	btnBatch := lcl.NewButton(mainPanel)
	btnBatch.SetParent(mainPanel)
	btnBatch.SetCaption("📊 批量发送测试")
	btnBatch.SetLeft(220)
	btnBatch.SetTop(row4Top)
	btnBatch.SetWidth(180)
	btnBatch.SetHint("连续发送5个通知测试")
	btnBatch.SetOnClick(func(sender lcl.IObject) {
		m.setStatus("正在批量发送...")

		go func() {
			for i := 1; i <= 5; i++ {
				opts := Options{
					ID:       fmt.Sprintf("batch-%d-%d", time.Now().Unix(), i),
					Title:    fmt.Sprintf("批量测试 #%d", i),
					Subtitle: "性能测试",
					Body:     fmt.Sprintf("这是第 %d 个测试通知", i),
				}

				m.notifService.SendNotification(opts)
				time.Sleep(500 * time.Millisecond)
			}

			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.setStatus("✅ 批量发送完成（5个通知）")
				m.appendLog("📊 批量发送完成: 5个通知\n")
			})
		}()
	})

	// 按钮14: 清空日志
	btnClearLog := lcl.NewButton(mainPanel)
	btnClearLog.SetParent(mainPanel)
	btnClearLog.SetCaption("📝 清空日志")
	btnClearLog.SetLeft(420)
	btnClearLog.SetTop(row4Top)
	btnClearLog.SetWidth(160)
	btnClearLog.SetHint("清空下方的日志区域")
	btnClearLog.SetOnClick(func(sender lcl.IObject) {
		m.logMemo.Clear()
		m.setStatus("日志已清空")
	})

	// ========== 日志区域 ==========
	logLabel := lcl.NewLabel(m)
	logLabel.SetParent(m)
	logLabel.SetCaption("📋 通知响应日志:")
	logLabel.SetLeft(20)
	logLabel.SetTop(290)

	m.logMemo = lcl.NewMemo(m)
	m.logMemo.SetParent(m)
	m.logMemo.SetLeft(20)
	m.logMemo.SetTop(315)
	m.logMemo.SetWidth(760)
	m.logMemo.SetHeight(280)
	m.logMemo.SetScrollBars(types.SsBoth)
	m.logMemo.SetReadOnly(true)

	// 初始日志
	m.appendLog("========================================\n")
	m.appendLog("  macOS 通知功能完整示例\n")
	m.appendLog("========================================\n\n")
	m.appendLog("使用说明:\n")
	m.appendLog("1. 首次使用请点击 '① 请求通知权限'\n")
	m.appendLog("2. 如果权限被拒绝，去 系统偏好设置 > 通知与焦点 中手动启用\n")
	m.appendLog("3. 应用必须打包为 .app 格式（包含 Info.plist 和 Bundle ID）\n")
	m.appendLog("4. 点击带操作的通知可以触发回调函数\n")
	m.appendLog("5. 所有通知响应会显示在此日志区域\n\n")
	m.appendLog("========================================\n\n")
}
