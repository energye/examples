#import <Cocoa/Cocoa.h>

#ifdef __cplusplus
extern "C" {
#endif

// 回调设置
typedef void (*ControlCallback)(const char *identifier, const char *value, const void *userData);

// 工具栏配置选项
typedef NS_OPTIONS(NSUInteger, ToolbarConfiguration) {
    ToolbarConfigurationNone = 0,
    ToolbarConfigurationAllowUserCustomization = 1 << 0,
    ToolbarConfigurationAutoSaveConfiguration = 1 << 1,
    ToolbarConfigurationShowSeparator = 1 << 2,
    ToolbarConfigurationDisplayModeIconOnly = 1 << 3,
    ToolbarConfigurationDisplayModeTextOnly = 1 << 4,
    ToolbarConfigurationDisplayModeIconAndText = 1 << 5
};

// 回调上下文结构体，替代全局回调
typedef struct {
    ControlCallback clickCallback;
    ControlCallback textChangedCallback;
    void *userData; // 用户自定义数据指针
} ToolbarCallbackContext;

// 控件样式结构体
typedef struct {
    CGFloat width;
    CGFloat height;
    NSBezelStyle bezelStyle;
    NSControlSize controlSize;
    NSFont *font;
    BOOL IsNavigational;
} ControlStyle;

// 动态添加控件
//void AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlStyle style, NSUInteger index);
void AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlStyle style);
void AddToolbarImageButton(unsigned long nsWindowHandle, const char *identifier, const char *imageName, const char *tooltip, ControlStyle style);
void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlStyle style);
void AddToolbarSearchField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlStyle style);
void AddToolbarCombobox(unsigned long nsWindowHandle, const char *identifier, const char **items, int count, ControlStyle style);
void AddToolbarCustomView(unsigned long nsWindowHandle, const char *identifier, ControlStyle style);

// 控件管理
const char *GetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier);
void SetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier, const char *value);
void SetToolbarControlEnabled(unsigned long nsWindowHandle, const char *identifier, bool enabled);

// 公共函数
ControlStyle CreateDefaultControlStyle();
ControlStyle CreateControlStyle(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font);
void ConfigureWindow(unsigned long nsWindowHandle, ToolbarConfiguration config, ToolbarCallbackContext callbackContext);

// 工具栏管理函数
void RemoveToolbarItem(unsigned long nsWindowHandle, const char *identifier);
void UpdateToolbarItemStyle(unsigned long nsWindowHandle, const char *identifier, ControlStyle style);
void InsertToolbarItemAtIndex(unsigned long nsWindowHandle, const char *identifier, int index);
void AddToolbarFlexibleSpace(unsigned long nsWindowHandle);
void AddToolbarSpace(unsigned long nsWindowHandle);
void AddToolbarSpaceByWidth(unsigned long nsWindowHandle, CGFloat width);
long GetToolbarItemCount(unsigned long nsWindowHandle);

#ifdef __cplusplus
}
#endif