#import "config.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// 创建工具栏事件回调上下文
ToolbarCallbackContext* CreateToolbarCallbackContext(long type, const NSString* identifier, const NSString* value, long index, void* owner, void* sender) {
    // 分配内存空间
    ToolbarCallbackContext* context = (ToolbarCallbackContext*)malloc(sizeof(ToolbarCallbackContext));
    if (!context) return NULL;  // 内存分配失败
    // 初始化基本字段
    context->type_ = type;
    context->index = index;
    context->owner = owner;
    context->sender = sender;
    // 深拷贝字符串字段
    context->identifier = identifier ? strdup([identifier UTF8String]) : strdup("");
    context->value = value ? strdup([value UTF8String]) : strdup("");
    return context;
}

// 释放工具栏事件回调上下文
void FreeToolbarCallbackContext(ToolbarCallbackContext* context) {
    if (!context) return;
    // 释放字符串内存
    free((void*)context->identifier);
    free((void*)context->value);
    // 释放结构体
    free(context);
}

static char kToolbarDelegateKey;

// 工具栏委托类
@interface MainToolbarDelegate : NSObject <NSToolbarDelegate, NSTextFieldDelegate, NSComboBoxDelegate, NSSearchFieldDelegate> {
    ControlEventCallback _callback;
    NSWindow *_window; // NSWindow
}

@property (nonatomic, strong) NSMutableDictionary<NSString *, NSView *> *controls;
@property (nonatomic, strong) NSMutableDictionary<NSString *, NSValue *> *controlProperty;
@property (nonatomic, strong) NSMutableArray<NSString *> *dynamicIdentifiers;
// @property (nonatomic, assign) ToolbarConfiguration configuration;

- (void)addControl:(NSView *)control forIdentifier:(NSString *)identifier withProperty:(ControlProperty)property;
- (NSView *)controlForIdentifier:(NSString *)identifier;
- (void)removeControlForIdentifier:(NSString *)identifier;
- (void)setCallback:(ControlEventCallback)callback;
- (void)setWindow:(NSWindow *)window;
- (NSWindow *)getWindow;
- (void)updateControlProperty:(NSString *)identifier withProperty:(ControlProperty)property;

- (void)windowDidResize:(NSNotification *)notification;
- (void)updateTextFieldWidthsForWindow:(NSWindow *)window;

@end

@implementation MainToolbarDelegate

- (instancetype)init {
    self = [super init];
    if (self) {
        self.controls = [NSMutableDictionary dictionary];
        self.dynamicIdentifiers = [NSMutableArray array];
        self.controlProperty = [NSMutableDictionary dictionary];
        _callback = NULL;
        // 监听窗口大小变化
        [[NSNotificationCenter defaultCenter] addObserver:self selector:@selector(windowDidResize:) name:NSWindowDidResizeNotification object:nil];
    }
    return self;
}

- (void)dealloc {
    NSLog(@"MainToolbarDelegate dealloc 释放");
    [[NSNotificationCenter defaultCenter] removeObserver:self];
    [super dealloc];
}


- (void)windowDidResize:(NSNotification *)notification {
    NSWindow *window = notification.object;
    // NSLog(@"windowDidResize");
    [self updateTextFieldWidthsForWindow:window];
}

- (void)updateTextFieldWidthsForWindow:(NSWindow *)window {
    CGFloat windowWidth = window.frame.size.width;
    CGFloat availableWidth = windowWidth - 180; // 减去交通灯区域和边距
}


- (void)addControl:(NSView *)control forIdentifier:(NSString *)identifier withProperty:(ControlProperty)property {
    NSLog(@"addControl");
    _controls[identifier] = control;
    // 存储控件样式
    NSValue *propertyValue = [NSValue value:&property withObjCType:@encode(ControlProperty)];
    _controlProperty[identifier] = propertyValue;
    if (![_dynamicIdentifiers containsObject:identifier]) {
        [_dynamicIdentifiers addObject:identifier];
    }
}

- (NSView *)controlForIdentifier:(NSString *)identifier {
    NSLog(@"controlForIdentifier");
    return _controls[identifier];
}

- (void)removeControlForIdentifier:(NSString *)identifier {
    NSLog(@"removeControlForIdentifier");
    // 从控件字典中移除
    [_controls removeObjectForKey:identifier];
    // 从样式字典中移除
    [_controlProperty removeObjectForKey:identifier];
    // 从标识符数组中移除
    [_dynamicIdentifiers removeObject:identifier];
}

- (void)setCallback:(ControlEventCallback)callback {
    _callback = callback;
}

- (void)setWindow:(NSWindow *)window {
    _window = window;
}

- (NSWindow *)getWindow {
  return _window;
}

- (void)updateControlProperty:(NSString *)identifier withProperty:(ControlProperty)property {
    NSLog(@"updateControlProperty");
    NSView *control = [self controlForIdentifier:identifier];
    if (!control) return;

    // 更新存储的样式
    NSValue *propertyValue = [NSValue value:&property withObjCType:@encode(ControlProperty)];
    _controlProperty[identifier] = propertyValue;

    // 应用样式到控件
    if ([control isKindOfClass:[NSControl class]]) {
        NSControl *ctrl = (NSControl *)control;
        ctrl.controlSize = property.controlSize;

        // 宽度约束
        if (property.width > 0) {
            // 移除现有宽度约束
            for (NSLayoutConstraint *constraint in control.constraints) {
                if (constraint.firstAttribute == NSLayoutAttributeWidth) {
                    [control removeConstraint:constraint];
                    break;
                }
            }
            // 添加新宽度约束
            [control.widthAnchor constraintEqualToConstant:property.width].active = YES;
        }

        // 高度约束
        if (property.height > 0) {
            // 移除现有高度约束
            for (NSLayoutConstraint *constraint in control.constraints) {
                if (constraint.firstAttribute == NSLayoutAttributeHeight) {
                    [control removeConstraint:constraint];
                    break;
                }
            }
            // 添加新高度约束
            [control.heightAnchor constraintEqualToConstant:property.height].active = YES;
        }

        // 特定控件类型的样式
        if ([control isKindOfClass:[NSButton class]]) {
            NSButton *button = (NSButton *)control;
            button.bezelStyle = property.bezelStyle;
            if (property.font) {
                button.font = property.font;
            }
        } else if ([control isKindOfClass:[NSTextField class]] ||
                   [control isKindOfClass:[NSSearchField class]] ||
                   [control isKindOfClass:[NSComboBox class]]) {
            NSTextField *textField = (NSTextField *)control;
            if (property.font) {
                textField.font = property.font;
            }
        }
    }
}

#pragma mark - Toolbar Delegate

- (NSArray<NSToolbarItemIdentifier> *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar {
    NSLog(@"toolbarDefaultItemIdentifiers");
    NSMutableArray *identifiers = [_dynamicIdentifiers copy];
    return identifiers;
}

- (NSArray<NSToolbarItemIdentifier> *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar {
    NSLog(@"toolbarAllowedItemIdentifiers");
    NSMutableArray *identifiers = [NSMutableArray arrayWithArray:_dynamicIdentifiers];
    // 添加系统标识符
    [identifiers addObject:NSToolbarFlexibleSpaceItemIdentifier];
    [identifiers addObject:NSToolbarSpaceItemIdentifier];
    return identifiers;
}

- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar
     itemForItemIdentifier:(NSToolbarItemIdentifier)itemIdentifier
 willBeInsertedIntoToolbar:(BOOL)flag {
    NSLog(@"toolbar");
    // 处理系统项
    if ([itemIdentifier isEqualToString:NSToolbarFlexibleSpaceItemIdentifier]) {
        return [[NSToolbarItem alloc] initWithItemIdentifier:NSToolbarFlexibleSpaceItemIdentifier];
    }
    if ([itemIdentifier isEqualToString:NSToolbarSpaceItemIdentifier]) {
        return [[NSToolbarItem alloc] initWithItemIdentifier:NSToolbarSpaceItemIdentifier];
    }
    // 处理动态控件
    NSView *control = [self controlForIdentifier:itemIdentifier];
    if (control) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:itemIdentifier];
        item.view = control;
        // 应用存储的样式
        NSValue *propertyValue = _controlProperty[itemIdentifier];
        if (propertyValue) {
            ControlProperty property;
            [propertyValue getValue:&property];
            item.navigational = property.IsNavigational; // 导航模式 靠左
            if (property.IsCenteredItem) {
                toolbar.centeredItemIdentifier = item.itemIdentifier;  // 设置为居中项
            }
            item.visibilityPriority = property.VisibilityPriority; // 可见优先级

            NSLog(@"toolbar %d %@ %d", property.IsNavigational, itemIdentifier, property.IsCenteredItem);

            [self updateControlProperty:itemIdentifier withProperty:property];
        }
        return item;
    }
    return nil;
}

#pragma mark - 事件处理

// 实现代理方法
- (void)searchFieldDidStartSearching:(NSSearchField *)sender {
    NSLog(@"搜索开始: %@", sender.stringValue);
    // 在这里处理搜索开始时的逻辑
}

- (void)searchFieldDidEndSearching:(NSSearchField *)sender {
    NSLog(@"搜索结束");
    // 在这里处理搜索结束时的逻辑
}

- (void)buttonClicked:(NSButton *)sender {
    NSLog(@"buttonClicked");
    if (_callback) {
        NSString *identifier = objc_getAssociatedObject(sender, @"identifier");
        if (identifier) {
            ToolbarCallbackContext *context = CreateToolbarCallbackContext(TCCClicked, identifier, @"", -1, _window, sender);
            @try{
                _callback(context);
            } @finally {
                FreeToolbarCallbackContext(context);
            }
        }
    }
}

- (void)comboBoxSelectionChanged:(NSComboBox *)sender {
    NSLog(@"comboBoxSelectionChanged");
    if (_callback) {
        NSString *identifier = objc_getAssociatedObject(sender, @"identifier");
        if (identifier) {
            NSInteger selectedIndex = [sender indexOfSelectedItem];
            ToolbarCallbackContext *context = CreateToolbarCallbackContext(TCCSelectionChanged, identifier, [sender stringValue], selectedIndex, _window, sender);
            @try{
                _callback(context);
            } @finally {
                FreeToolbarCallbackContext(context);
            }
        }
    }
}

// 用户选择发生变化时触发
- (void)comboBoxSelectionDidChange:(NSNotification *)notification {
    NSLog(@"comboBoxSelectionChanged");
    if (_callback) {
        id control = notification.object;
        NSString *identifier = objc_getAssociatedObject(control, @"identifier");
        if (identifier) {
            NSInteger selectedIndex = [control indexOfSelectedItem];
            ToolbarCallbackContext *context = CreateToolbarCallbackContext(TCCSelectionDidChange, identifier, [control stringValue], selectedIndex, _window, control);
            @try{
                _callback(context);
            } @finally {
                FreeToolbarCallbackContext(context);
            }
        }
    }
}


- (void)controlTextDidChange:(NSNotification *)notification {
    NSLog(@"controlTextDidChange");
    if (_callback) {
        id control = notification.object;
        NSString *identifier = objc_getAssociatedObject(control, @"identifier");
        if (identifier) {
            ToolbarCallbackContext *context = CreateToolbarCallbackContext(TCCTextDidChange, identifier, [control stringValue], -1, _window, control);
            @try{
                _callback(context);
            } @finally {
                FreeToolbarCallbackContext(context);
            }
        }
    }
}

- (void)controlTextDidEndEditing:(NSNotification *)notification {
    NSLog(@"controlTextDidEndEditing");
    if (_callback) {
        id control = notification.object;
        NSString *identifier = objc_getAssociatedObject(control, @"identifier");
        if (identifier) {
            ToolbarCallbackContext *context = CreateToolbarCallbackContext(TCCTextDidEndEditing, identifier, [control stringValue], -1, _window, control);
            @try{
                _callback(context);
            } @finally {
                FreeToolbarCallbackContext(context);
            }
        }
    }
}

@end

#pragma mark - 公共函数实现

// 创建默认控件样式
ControlProperty CreateDefaultControlProperty() {
    ControlProperty property;
    property.width = 0; // 0表示自动大小
    property.height = 0;
    property.minWidth = 0;
    property.maxWidth = 0;
    property.bezelStyle = NSBezelStyleTexturedRounded;
    property.controlSize = NSControlSizeRegular;
    property.font = nil;
    property.VisibilityPriority = NSToolbarItemVisibilityPriorityStandard;
    return property;
}

// 创建自定义控件样式
ControlProperty CreateControlProperty(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font) {
    ControlProperty property;
    property.width = width;
    property.height = height;
    property.bezelStyle = bezelStyle;
    property.controlSize = controlSize;
    property.font = (__bridge NSFont *)font;
    return property;
}

// 初始化函数
__attribute__((constructor))
static void initializeDelegateMap() {
    // NSLog(@"initializeDelegateMap");
}

// 设置窗口背景色
void SetWindowBackgroundColor(unsigned long nsWindowHandle, Color color) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    NSColor *bgColor = [NSColor colorWithCalibratedRed:color.Red
                                                 green:color.Green
                                                  blue:color.Blue
                                                 alpha:color.Alpha];
    window.backgroundColor = bgColor;
//    NSView *contentView = window.contentView;
//    contentView.wantsLayer = YES;
//    contentView.layer.backgroundColor = bgColor.CGColor;
}

// 配置窗口
void CreateToolbar(unsigned long nsWindowHandle, ToolbarConfiguration config, ControlEventCallback callback, void **outToolbarDelegate, void** outToolbar) {
    NSLog(@"CreateToolbar");
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;

    // 创建工具栏代理
    MainToolbarDelegate *toolbarDelegate = [[MainToolbarDelegate alloc] init];
    [toolbarDelegate setCallback:callback];
    [toolbarDelegate setWindow:window];

    // 创建工具栏
    NSToolbar *toolbar = [[NSToolbar alloc] initWithIdentifier:@"ENERGY.ToolBar"];
    toolbar.delegate = toolbarDelegate;
    // 设置显示模式
    window.titlebarAppearsTransparent = config.Transparent;

    window.showsToolbarButton = config.ShowsToolbarButton;
    window.toolbarStyle = config.Style;
    window.titlebarSeparatorStyle = config.SeparatorStyle;
    toolbar.allowsUserCustomization = config.IsAllowsUserCustomization;
    toolbar.autosavesConfiguration = config.IsAutoSavesConfiguration;
    toolbar.displayMode = config.DisplayMode;
    toolbar.sizeMode = config.SizeMode; //NSToolbarSizeModeRegular; // 或 NSToolbarSizeModeSmall

    window.toolbar = toolbar;

    // 保留委托对象
    objc_setAssociatedObject(window, &kToolbarDelegateKey, toolbarDelegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    if (outToolbarDelegate) {
        *outToolbarDelegate = (__bridge void*)(toolbarDelegate);
    }
    if (outToolbar) {
        *outToolbar = (__bridge void*)(toolbar);
    }
}

// 向 toolbar 添加控件
void ToolbarAddControl(void* nsDelegate, void* nsToolbar, void* nsControl, const char *identifier, ControlProperty property) {
    if (!nsDelegate || !nsToolbar || !nsControl || !identifier) {
        NSLog(@"[ERROR] AddToolbarControl 必要入参为空");
        return;
    }
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSToolbar *toolbar = (NSToolbar*)nsToolbar;
    NSView *view = (NSView*)nsControl;
    NSString *idStr = [NSString stringWithUTF8String:identifier];
    if (!toolbar || !delegate || !view || !idStr) {
        NSLog(@"[ERROR] AddToolbarControl 必要转换参数为空");
        return;
    }
    // 添加到委托 维护, 工具栏获取时使用
    [delegate addControl:view forIdentifier:idStr withProperty:property];
    // 添加到工具栏
    [toolbar insertItemWithItemIdentifier:idStr atIndex:toolbar.items.count];
}

#pragma mark - 动态控件创建函数

// 通用函数：通过NSControl设置控件属性（适用于按钮、文本框等）
void ConfigureControl(NSControl *control, NSString *tooltipStr, ControlProperty property) {
    control.controlSize = property.controlSize;
    if (tooltipStr) {
        control.toolTip = tooltipStr;
    }
    if (property.font) {
        control.font = property.font;
    }
    if (property.width > 0) {
        [control.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [control.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }
    if (property.minWidth > 0) {
        [control.widthAnchor constraintGreaterThanOrEqualToConstant:property.minWidth].active = YES;
    }
    if (property.maxWidth > 0) {
        [control.widthAnchor constraintLessThanOrEqualToConstant:property.maxWidth].active = YES;
    }
}

void* NewButton(void* nsDelegate, const char *identifier, const char *title, const char *tooltip, ControlProperty property) {
    if (!title) {
        NSLog(@"[ERROR] NewButton 必要参数为空");
        return nil;
    }
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *titleStr = [NSString stringWithUTF8String:title];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    NSButton *button = [NSButton buttonWithTitle:titleStr target:delegate action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);
    ConfigureControl(button, tooltipStr, property);
    return (__bridge void*)(button);
}

void* NewImageButton(void* nsDelegate, const char *identifier, const char *image, const char *tooltip, ControlProperty property) {
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *imageNameStr = [NSString stringWithUTF8String:image];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    NSButton *button = [NSButton buttonWithImage:[NSImage imageWithSystemSymbolName:imageNameStr accessibilityDescription:nil]
                                                                             target:delegate
                                                                             action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    button.imagePosition = NSImageOnly;
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);
    ConfigureControl(button, tooltipStr, property);
    return button;
}

void* AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);
    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *titleStr = [NSString stringWithUTF8String:title];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    // 创建按钮
    NSButton *button = [NSButton buttonWithTitle:titleStr target:delegate action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    button.controlSize = property.controlSize;
    if (tooltipStr) {
        button.toolTip = tooltipStr;
    }
    if (property.font) {
        button.font = property.font;
    }
    // 设置尺寸约束
    if (property.width > 0) {
        [button.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [button.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }
    // 关联标识符
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);
    // 添加到委托
    [delegate addControl:button forIdentifier:idStr withProperty:property];
    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
    return (__bridge void*)(button);
}

void AddToolbarImageButton(unsigned long nsWindowHandle, const char *identifier, const char *imageName, const char *tooltip, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *imageNameStr = [NSString stringWithUTF8String:imageName];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;

    NSLog(@"Loading toolbar image: %@", imageNameStr);

    // 创建图片按钮
//     NSButton *button = [NSButton buttonWithImage:[NSImage imageNamed:imageNameStr]
//                                          target:delegate
//                                          action:@selector(buttonClicked:)];
    NSButton *button = [NSButton buttonWithImage:[NSImage imageWithSystemSymbolName:imageNameStr accessibilityDescription:nil]
                                         target:delegate
                                         action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    button.controlSize = property.controlSize;
    button.imagePosition = NSImageOnly;
    if (tooltipStr) {
        button.toolTip = tooltipStr;
    }
    if (property.font) {
        button.font = property.font;
    }

    // 设置尺寸约束
    if (property.width > 0) {
        [button.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [button.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:button forIdentifier:idStr withProperty:property];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *placeholderStr = placeholder ? [NSString stringWithUTF8String:placeholder] : nil;

    // 创建文本框
    NSTextField *textField = [[NSTextField alloc] init];
    textField.placeholderString = placeholderStr;
    textField.delegate = delegate;
    textField.controlSize = property.controlSize;

    // textField.alignment = NSTextAlignmentCenter;    // 设置水平居中

    if (property.font) {
        textField.font = property.font;
    }

    // 设置自动调整大小的属性
    [textField setContentHuggingPriority:NSLayoutPriorityDefaultLow
                          forOrientation:NSLayoutConstraintOrientationHorizontal];
    [textField setContentCompressionResistancePriority:NSLayoutPriorityDefaultLow
                                        forOrientation:NSLayoutConstraintOrientationHorizontal];

    // 设置尺寸约束
    if (property.width > 0) {
        [textField.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [textField.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }
    if (property.minWidth > 0) {
        [textField.widthAnchor constraintGreaterThanOrEqualToConstant:property.minWidth].active = YES;
    }
    if (property.maxWidth > 0) {
        [textField.widthAnchor constraintLessThanOrEqualToConstant:property.maxWidth].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(textField, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:textField forIdentifier:idStr withProperty:property];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void* AddToolbarSearchField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *placeholderStr = placeholder ? [NSString stringWithUTF8String:placeholder] : nil;

    // 创建搜索框
    NSSearchField *searchField = [[NSSearchField alloc] init];
    searchField.placeholderString = placeholderStr;
    searchField.delegate = delegate;
    searchField.controlSize = property.controlSize;
    if (property.font) {
        searchField.font = property.font;
    }

    // 设置尺寸约束
    if (property.width > 0) {
        [searchField.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [searchField.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }
    // 最小和最大宽度约束
    if (property.minWidth > 0) {
        NSLayoutConstraint *minWidthConstraint = [searchField.widthAnchor constraintGreaterThanOrEqualToConstant:property.minWidth];
        minWidthConstraint.priority = NSLayoutPriorityDefaultHigh;
        minWidthConstraint.active = YES;
    }
    if (property.maxWidth > 0) {
        NSLayoutConstraint *maxWidthConstraint = [searchField.widthAnchor constraintLessThanOrEqualToConstant:property.maxWidth];
        maxWidthConstraint.priority = NSLayoutPriorityDefaultHigh;
        maxWidthConstraint.active = YES;
    }
    [searchField setContentHuggingPriority:NSLayoutPriorityDefaultLow
                          forOrientation:NSLayoutConstraintOrientationHorizontal];
    [searchField setContentCompressionResistancePriority:NSLayoutPriorityDefaultLow
                                            forOrientation:NSLayoutConstraintOrientationHorizontal];

    objc_setAssociatedObject(searchField, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);// 关联标识符
    [delegate addControl:searchField forIdentifier:idStr withProperty:property];// 添加到委托
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count]; // 添加到工具栏
//     [window layoutIfNeeded];
    return (__bridge void*)(searchField);
}

void AddToolbarCombobox(unsigned long nsWindowHandle, const char *identifier, const char **items, int count, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 创建下拉框
    NSComboBox *comboBox = [[NSComboBox alloc] init];
    comboBox.delegate = delegate;
    comboBox.controlSize = property.controlSize;
    [comboBox setEditable:NO];
    if (property.font) {
        comboBox.font = property.font;
    }

    // 添加选项
    for (int i = 0; i < count; i++) {
        [comboBox addItemWithObjectValue:[NSString stringWithUTF8String:items[i]]];
    }

    // 设置默认选择
    if (count > 0) {
        [comboBox selectItemAtIndex:0];
    }

    // 设置尺寸约束
    if (property.width > 0) {
        [comboBox.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [comboBox.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(comboBox, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:comboBox forIdentifier:idStr withProperty:property];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarCustomView(unsigned long nsWindowHandle, const char *identifier, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 创建自定义容器
    NSView *container = [[NSView alloc] init];

    // 设置尺寸约束
    if (property.width > 0) {
        [container.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [container.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(container, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:container forIdentifier:idStr withProperty:property];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

#pragma mark - 工具栏管理函数

void RemoveToolbarItem(unsigned long nsWindowHandle, const char *identifier) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 从委托中移除控件
    [delegate removeControlForIdentifier:idStr];

    // 从工具栏中移除项
    NSUInteger index = [window.toolbar.items indexOfObjectPassingTest:^BOOL(NSToolbarItem * _Nonnull obj, NSUInteger idx, BOOL * _Nonnull stop) {
        return [obj.itemIdentifier isEqualToString:idStr];
    }];

    if (index != NSNotFound) {
        [window.toolbar removeItemAtIndex:index];
    }
}

void UpdateToolbarItemProperty(unsigned long nsWindowHandle, const char *identifier, ControlProperty property) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    [delegate updateControlProperty:idStr withProperty:property];
}

void InsertToolbarItemAtIndex(unsigned long nsWindowHandle, const char *identifier, int index) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 确保索引在有效范围内
    NSUInteger itemCount = window.toolbar.items.count;
    NSUInteger insertIndex = MIN(MAX(index, 0), itemCount);

    // 从当前位置移除（如果存在）
    NSUInteger currentIndex = [window.toolbar.items indexOfObjectPassingTest:^BOOL(NSToolbarItem * _Nonnull obj, NSUInteger idx, BOOL * _Nonnull stop) {
        return [obj.itemIdentifier isEqualToString:idStr];
    }];

    if (currentIndex != NSNotFound) {
        [window.toolbar removeItemAtIndex:currentIndex];
        // 如果当前索引在插入索引之前，需要调整插入索引
        if (currentIndex < insertIndex) {
            insertIndex--;
        }
    }

    // 插入到新位置
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:insertIndex];
}

void AddToolbarFlexibleSpace(unsigned long nsWindowHandle) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);
    if (!delegate) return;
    NSString *flexSpaceId = NSToolbarFlexibleSpaceItemIdentifier;
    // 添加到委托（使用nil控件，因为这是系统项）
//     ControlProperty property = CreateDefaultControlProperty();
//     [delegate addControl:nil forIdentifier:flexSpaceId withProperty:property];
 // 确保灵活空间标识符在动态标识符列表中
//     if (![delegate.dynamicIdentifiers containsObject:flexSpaceId]) {
//         [delegate.dynamicIdentifiers addObject:flexSpaceId];
//     }
    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:flexSpaceId atIndex:window.toolbar.items.count];
}

void AddToolbarSpace(unsigned long nsWindowHandle) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);
    if (!delegate) return;
    NSString *spaceId = NSToolbarSpaceItemIdentifier;
    // 添加到委托（使用nil控件，因为这是系统项）
//     ControlProperty property = CreateDefaultControlProperty();
//     [delegate addControl:nil forIdentifier:spaceId withProperty:property];
    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:spaceId atIndex:window.toolbar.items.count];
}

void AddToolbarSpaceByWidth(unsigned long nsWindowHandle, CGFloat width) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);

    // 创建固定空格标识符
    NSString *spaceIdentifier = [NSString stringWithFormat:@"FixedSpace_%.0f", width];

    // 创建固定宽度的视图
    NSView *spaceView = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, width, 1)];
    spaceView.translatesAutoresizingMaskIntoConstraints = NO;  // 关闭自动尺寸调整
    [spaceView.widthAnchor constraintEqualToConstant:width].active = YES;

    // 添加到委托
    [delegate addControl:spaceView forIdentifier:spaceIdentifier withProperty:CreateDefaultControlProperty()];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:spaceIdentifier atIndex:window.toolbar.items.count];
}

long GetToolbarItemCount(unsigned long nsWindowHandle) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    return window.toolbar.items.count;
}

// 循环工具栏每项获取 NSControl，通过代理获取有问题啊。
NSView* GetToolbarControl(unsigned long nsWindowHandle, const char *identifier) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 使用代理获取 controls
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, &kToolbarDelegateKey);
    NSView *control = [delegate controlForIdentifier:idStr];
    if (!control) return nil;
    return control;

    // 使用循环获取 controls
//     NSToolbar *toolbar = window.toolbar;
//     if (![toolbar isKindOfClass:[NSToolbar class]]) {
//         NSLog(@"GetToolbarControl not kind NSToolbar class");
//         return nil;
//     }
//     for (NSToolbarItem *item in toolbar.items) {
//         if (![item.itemIdentifier isEqualToString:idStr]) continue;
//         return item.view;
//     }
//     return nil;
}

#pragma mark - 控件管理函数

const char *GetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier) {
    NSView *control = GetToolbarControl(nsWindowHandle, identifier);
    if (!control) return NULL;
    NSString *idStr = [NSString stringWithUTF8String:identifier];

    if ([control isKindOfClass:[NSTextField class]]) {
        return [[(NSTextField *)control stringValue] UTF8String];
    }
    else if ([control isKindOfClass:[NSComboBox class]]) {
        return [[(NSComboBox *)control stringValue] UTF8String];
    }
    else if ([control isKindOfClass:[NSSearchField class]]) {
        return [[(NSSearchField *)control stringValue] UTF8String];
    }
    return NULL;
}

void SetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier, const char *value) {
    NSView *control = GetToolbarControl(nsWindowHandle, identifier);
    if (!control) return;
    NSString *valueStr = [NSString stringWithUTF8String:value];
    if ([control isKindOfClass:[NSTextField class]]) {
        [(NSTextField *)control setStringValue:valueStr];
    }
    else if ([control isKindOfClass:[NSComboBox class]]) {
        [(NSComboBox *)control setStringValue:valueStr];
    }
    else if ([control isKindOfClass:[NSSearchField class]]) {
        [(NSSearchField *)control setStringValue:valueStr];
    }
}

void SetToolbarControlEnabled(unsigned long nsWindowHandle, const char *identifier, bool enabled) {
    NSView *control = GetToolbarControl(nsWindowHandle, identifier);
    if (!control) return;
    if ([control isKindOfClass:[NSControl class]]) {
        [(NSControl *)control setEnabled:(BOOL)enabled];
    }
}

void SetToolbarControlHidden(unsigned long nsWindowHandle, const char *identifier, bool hidden) {
    NSView *control = GetToolbarControl(nsWindowHandle, identifier);
    if (!control) {
        NSLog(@"获取 NSView(control)失败");
        return;
    }
    if ([control isKindOfClass:[NSControl class]]) {
        [(NSControl *)control setHidden:(BOOL)hidden];
    }
}

// 通过指针获取搜索框的值
const char* GetSearchFieldText(void* ptr) {
    NSSearchField* searchField = (__bridge NSSearchField*)(ptr);
    NSString* nsText = [searchField stringValue];
    // 转换为 C 字符串（需注意：返回的指针需在 Go 中及时处理，避免被释放）
    return [nsText UTF8String];
}

// 通过指针设置搜索框文本
void SetSearchFieldText(void* ptr, const char* text) {
    NSSearchField* searchField = (__bridge NSSearchField*)(ptr);
    NSString* nsText = [NSString stringWithUTF8String:text];
    [searchField setStringValue:nsText];
}

// 通过指针设置搜索框文本
void UpdateSearchFieldWidth(void* ptr, CGFloat width) {
    NSSearchField* searchField = (__bridge NSSearchField*)(ptr);
    // 移除现有宽度约束
    for (NSLayoutConstraint *constraint in searchField.constraints) {
        if (constraint.firstAttribute == NSLayoutAttributeWidth) {
            [searchField removeConstraint:constraint];
            break;
        }
    }
    // 添加新宽度约束并设置高优先级
    NSLayoutConstraint *widthConstraint = [searchField.widthAnchor constraintEqualToConstant:width];
    widthConstraint.priority = NSLayoutPriorityRequired;
    widthConstraint.active = YES;
}
