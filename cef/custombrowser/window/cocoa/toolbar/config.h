#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

#ifdef __cplusplus
extern "C" {
#endif

// 回调设置
typedef void (*ControlCallback)(const char *identifier, const char *value, const void *userData);

enum {
    TCCClicked = 1,
    TCCTextDidChange = 2,
    TCCTextDidEndEditing = 3,
    TCCSelectionChanged = 4,
    TCCSelectionDidChange = 5
};

// 通用事件回调事件参数
typedef struct {
    long type_; // 1: 点击事件 2: 文本改变事件 3:文本提交事件 4:下拉框回车/离开焦点事件 5:下拉框选择事件
    const char *identifier; // 控件标识
    const char *value; // 控件值
    long index; // 值索引
    void *owner; // 控件所属对象
    void *sender; // 控件
} ToolbarCallbackContext;

// 通用事件回调事件类型
typedef void (*ControlEventCallback)(ToolbarCallbackContext *context);
// 创建事件对象
ToolbarCallbackContext* CreateToolbarCallbackContext(long type, const NSString* identifier, const NSString* value, long index, void* owner, void* sender);
// 释放事件对象
void FreeToolbarCallbackContext(ToolbarCallbackContext* context);

// 工具栏配置选项
typedef struct {
    BOOL            IsAllowsUserCustomization;
    BOOL            IsAutoSavesConfiguration;
	BOOL            Transparent;
	BOOL            ShowsToolbarButton;
	NSUInteger      SeparatorStyle;
    NSUInteger      DisplayMode;
    NSUInteger      SizeMode;
    NSUInteger      Style;
} ToolbarConfiguration;

// 控件样式结构体
typedef struct {
    CGFloat         width;
    CGFloat         height;
    CGFloat         minWidth;
    CGFloat         maxWidth;
    NSBezelStyle    bezelStyle;
    NSControlSize   controlSize;
    NSFont          *font;
    BOOL            IsNavigational;
    BOOL            IsCenteredItem;
    NSInteger       VisibilityPriority;
} ControlProperty;

// 动态添加控件
void AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlProperty property);
void AddToolbarImageButton(unsigned long nsWindowHandle, const char *identifier, const char *iconName, const char *tooltip, ControlProperty property);
void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlProperty property);
void* AddToolbarSearchField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlProperty property);
void AddToolbarCombobox(unsigned long nsWindowHandle, const char *identifier, const char **items, int count, ControlProperty property);
void AddToolbarCustomView(unsigned long nsWindowHandle, const char *identifier, ControlProperty property);

// 控件管理
const char *GetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier);
void SetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier, const char *value);
void SetToolbarControlEnabled(unsigned long nsWindowHandle, const char *identifier, bool enabled);
void SetToolbarControlHidden(unsigned long nsWindowHandle, const char *identifier, bool hidden);
const char* GetSearchFieldText(void* searchFieldPtr);
void SetSearchFieldText(void* ptr, const char* text);
void UpdateSearchFieldWidth(void* ptr, CGFloat width);

// 公共函数
ControlProperty CreateDefaultControlProperty();
ControlProperty CreateControlProperty(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font);
void ConfigureWindow(unsigned long nsWindowHandle, ToolbarConfiguration config, ControlEventCallback callback, void *owner);

// 工具栏管理函数
void RemoveToolbarItem(unsigned long nsWindowHandle, const char *identifier);
void UpdateToolbarItemProperty(unsigned long nsWindowHandle, const char *identifier, ControlProperty property);
void InsertToolbarItemAtIndex(unsigned long nsWindowHandle, const char *identifier, int index);
void AddToolbarFlexibleSpace(unsigned long nsWindowHandle);
void AddToolbarSpace(unsigned long nsWindowHandle);
void AddToolbarSpaceByWidth(unsigned long nsWindowHandle, CGFloat width);
long GetToolbarItemCount(unsigned long nsWindowHandle);


#ifdef __cplusplus
}
#endif