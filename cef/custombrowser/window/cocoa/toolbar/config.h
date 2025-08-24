#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>
#import <go_data.h>

#ifdef __cplusplus
extern "C" {
#endif


const char* GetStringConstValue(const void* str);


// 颜色
typedef struct {
    CGFloat Red;
    CGFloat Green;
    CGFloat Blue;
    CGFloat Alpha;
}  Color;


// 通用事件回调事件参数
typedef struct {
    const   char *identifier; // 控件唯一标识
    const   char *value; // 控件值
    long    index; // 值索引
    void    *owner; // 控件所属对象
    void    *sender; // 控件
    GoData  *inputData;// 传入数据
} ToolbarCallbackContext;

// 通用事件回调事件类型
typedef GoData* (*ControlEventCallback)(ToolbarCallbackContext *context);
// 创建事件对象
ToolbarCallbackContext* CreateToolbarCallbackContext(const NSString* identifier, const NSString* value, long index, void* owner, void* sender);
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
void* AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlProperty property);
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
NSImage* imageFromBytes(const uint8_t* data, size_t length);
ControlProperty CreateDefaultControlProperty();
ControlProperty CreateControlProperty(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font);

// 设置窗口背景色
void SetWindowBackgroundColor(unsigned long nsWindowHandle, Color color);

// 工具栏
void CreateToolbar(unsigned long nsWindowHandle, ToolbarConfiguration config, ControlEventCallback callback, void **outToolbarDelegate, void** outToolbar);
void ToolbarAddControl(void* nsDelegate, void* nsToolbar, void* nsControl, const char *identifier, ControlProperty property);

// 配置控件所属通用父类为 NSControl 属性
void ConfigureControl(NSControl *control, NSString *tooltipStr, ControlProperty property);
// 控件创建 Button
void* NewButton(void* delegate, const char *identifier, const char *title, const char *tooltip, ControlProperty property);
void* NewImageButtonFormImage(void* nsDelegate, const char *identifier, const char *image, const char *tooltip, ControlProperty property);
void* NewImageButtonFormBytes(void* nsDelegate, const char *identifier, const uint8_t* data, size_t length, const char *tooltip, ControlProperty property);
// 控件创建 TextField
void* NewTextField(void* nsDelegate, const char *identifier, const char *placeholder, const char *tooltip, ControlProperty property) ;
// Space
void AddToolbarFlexibleSpace(void* nsToolbar);
void AddToolbarSpace(void* nsToolbar);
//void AddToolbarSpaceByWidth(void* nsToolbar, CGFloat width);

// 工具栏管理函数
void RemoveToolbarItem(unsigned long nsWindowHandle, const char *identifier);
void UpdateToolbarItemProperty(unsigned long nsWindowHandle, const char *identifier, ControlProperty property);
void InsertToolbarItemAtIndex(unsigned long nsWindowHandle, const char *identifier, int index);
long GetToolbarItemCount(unsigned long nsWindowHandle);

// 工具栏委托类
@interface MainToolbarDelegate : NSObject <NSToolbarDelegate, NSTextFieldDelegate, NSComboBoxDelegate, NSSearchFieldDelegate> {
    ControlEventCallback _callback;
    NSWindow *_window; // NSWindow
    NSToolbar *_toolbar;
}

@property (nonatomic, strong) NSMutableDictionary<NSString *, NSView *> *controls;
@property (nonatomic, strong) NSMutableDictionary<NSString *, NSValue *> *controlProperty;
@property (nonatomic, strong) NSMutableArray<NSString *> *dynamicIdentifiers;

- (void)addControl:(NSView *)control forIdentifier:(NSString *)identifier withProperty:(ControlProperty)property;
- (NSView *)controlForIdentifier:(NSString *)identifier;
- (void)removeControlForIdentifier:(NSString *)identifier;
- (void)setCallback:(ControlEventCallback)callback withWindow:(NSWindow *)window withToolbar:(NSToolbar *)toolbar;
- (void)updateControlProperty:(NSString *)identifier withProperty:(ControlProperty)property;

@end


#ifdef __cplusplus
}
#endif