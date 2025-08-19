#import "window_config_darwin.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// 工具栏委托类
@interface MainToolbarDelegate : NSObject <NSToolbarDelegate, NSTextFieldDelegate, NSComboBoxDelegate, NSSearchFieldDelegate> {
    NSMutableDictionary<NSString *, NSView *> *_controls;
    NSMutableArray<NSString *> *_dynamicIdentifiers;
    NSMutableDictionary<NSString *, NSValue *> *_controlStyles;
    ToolbarCallbackContext _callbackContext;
}

@property (nonatomic, assign) ToolbarConfiguration configuration;

- (void)addControl:(NSView *)control forIdentifier:(NSString *)identifier withStyle:(ControlStyle)style;
- (NSView *)controlForIdentifier:(NSString *)identifier;
- (void)removeControlForIdentifier:(NSString *)identifier;
- (void)setCallbackContext:(ToolbarCallbackContext)context;
- (void)updateControlStyle:(NSString *)identifier withStyle:(ControlStyle)style;

@end

@implementation MainToolbarDelegate

- (instancetype)init {
    self = [super init];
    if (self) {
        _controls = [NSMutableDictionary dictionary];
        _dynamicIdentifiers = [NSMutableArray array];
        _controlStyles = [NSMutableDictionary dictionary];
        _callbackContext.clickCallback = NULL;
        _callbackContext.textChangedCallback = NULL;
        _callbackContext.userData = NULL;
        _configuration = ToolbarConfigurationNone;
    }
    return self;
}

- (void)addControl:(NSView *)control forIdentifier:(NSString *)identifier withStyle:(ControlStyle)style {
    _controls[identifier] = control;
    // 存储控件样式
    NSValue *styleValue = [NSValue value:&style withObjCType:@encode(ControlStyle)];
    _controlStyles[identifier] = styleValue;

    if (![_dynamicIdentifiers containsObject:identifier]) {
        [_dynamicIdentifiers addObject:identifier];
    }
}

- (NSView *)controlForIdentifier:(NSString *)identifier {
    return _controls[identifier];
}

- (void)removeControlForIdentifier:(NSString *)identifier {
    // 从控件字典中移除
    [_controls removeObjectForKey:identifier];
    // 从样式字典中移除
    [_controlStyles removeObjectForKey:identifier];
    // 从标识符数组中移除
    [_dynamicIdentifiers removeObject:identifier];
}

- (void)setCallbackContext:(ToolbarCallbackContext)context {
    _callbackContext = context;
}

- (void)updateControlStyle:(NSString *)identifier withStyle:(ControlStyle)style {
    NSView *control = [self controlForIdentifier:identifier];
    if (!control) return;

    // 更新存储的样式
    NSValue *styleValue = [NSValue value:&style withObjCType:@encode(ControlStyle)];
    _controlStyles[identifier] = styleValue;

    // 应用样式到控件
    if ([control isKindOfClass:[NSControl class]]) {
        NSControl *ctrl = (NSControl *)control;
        ctrl.controlSize = style.controlSize;

        // 宽度约束
        if (style.width > 0) {
            // 移除现有宽度约束
            for (NSLayoutConstraint *constraint in control.constraints) {
                if (constraint.firstAttribute == NSLayoutAttributeWidth) {
                    [control removeConstraint:constraint];
                    break;
                }
            }
            // 添加新宽度约束
            [control.widthAnchor constraintEqualToConstant:style.width].active = YES;
        }

        // 高度约束
        if (style.height > 0) {
            // 移除现有高度约束
            for (NSLayoutConstraint *constraint in control.constraints) {
                if (constraint.firstAttribute == NSLayoutAttributeHeight) {
                    [control removeConstraint:constraint];
                    break;
                }
            }
            // 添加新高度约束
            [control.heightAnchor constraintEqualToConstant:style.height].active = YES;
        }

        // 特定控件类型的样式
        if ([control isKindOfClass:[NSButton class]]) {
            NSButton *button = (NSButton *)control;
            button.bezelStyle = style.bezelStyle;
            if (style.font) {
                button.font = style.font;
            }
        } else if ([control isKindOfClass:[NSTextField class]] ||
                   [control isKindOfClass:[NSSearchField class]] ||
                   [control isKindOfClass:[NSComboBox class]]) {
            NSTextField *textField = (NSTextField *)control;
            if (style.font) {
                textField.font = style.font;
            }
        }
    }
}

#pragma mark - Toolbar Delegate

- (NSArray<NSToolbarItemIdentifier> *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar {
    return [_dynamicIdentifiers copy];
}

- (NSArray<NSToolbarItemIdentifier> *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar {
    NSMutableArray *identifiers = [NSMutableArray arrayWithArray:_dynamicIdentifiers];
    return identifiers;
}

- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar
     itemForItemIdentifier:(NSToolbarItemIdentifier)itemIdentifier
 willBeInsertedIntoToolbar:(BOOL)flag {

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
        NSValue *styleValue = _controlStyles[itemIdentifier];
        if (styleValue) {
            ControlStyle style;
            [styleValue getValue:&style];
            [self updateControlStyle:itemIdentifier withStyle:style];
        }

        return item;
    }

    return nil;
}

#pragma mark - 事件处理

- (void)buttonClicked:(NSButton *)sender {
    if (_callbackContext.clickCallback) {
        NSString *identifier = objc_getAssociatedObject(sender, @"identifier");
        if (identifier) {
            _callbackContext.clickCallback([identifier UTF8String], "", _callbackContext.userData);
        }
    }
}

- (void)comboBoxSelectionChanged:(NSComboBox *)sender {
    if (_callbackContext.clickCallback) {
        NSString *identifier = objc_getAssociatedObject(sender, @"identifier");
        if (identifier) {
            _callbackContext.clickCallback([identifier UTF8String], [[sender stringValue] UTF8String], _callbackContext.userData);
        }
    }
}

- (void)controlTextDidChange:(NSNotification *)notification {
    if (_callbackContext.textChangedCallback) {
        id control = notification.object;
        NSString *identifier = objc_getAssociatedObject(control, @"identifier");
        if (identifier) {
            NSString *value = [control stringValue];
            _callbackContext.textChangedCallback([identifier UTF8String], [value UTF8String], _callbackContext.userData);
        }
    }
}

@end

#pragma mark - 公共函数实现

// 创建默认控件样式
ControlStyle CreateDefaultControlStyle() {
    ControlStyle style;
    style.width = 0; // 0表示自动大小
    style.height = 0;
    style.bezelStyle = NSBezelStyleTexturedRounded;
    style.controlSize = NSControlSizeRegular;
    style.font = nil;
    return style;
}

// 创建自定义控件样式
ControlStyle CreateControlStyle(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font) {
    ControlStyle style;
    style.width = width;
    style.height = height;
    style.bezelStyle = bezelStyle;
    style.controlSize = controlSize;
    style.font = (__bridge NSFont *)font;
    return style;
}

// 获取窗口指针
void *GetNSWindowFromNSView(unsigned long nsViewHandle) {
    NSView *view = (__bridge NSView *)(void *)nsViewHandle;
    return (__bridge void *)[view window];
}

// 配置窗口
void ConfigureWindow(unsigned long nsWindowHandle, ToolbarConfiguration config, ToolbarCallbackContext callbackContext) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;

    // 创建工具栏
    MainToolbarDelegate *toolbarDelegate = [[MainToolbarDelegate alloc] init];
    toolbarDelegate.configuration = config;
    [toolbarDelegate setCallbackContext:callbackContext];

    NSToolbar *toolbar = [[NSToolbar alloc] initWithIdentifier:@"MainIDE.ToolBar"];
    toolbar.delegate = toolbarDelegate;
    toolbar.allowsUserCustomization = (config & ToolbarConfigurationAllowUserCustomization) != 0;
    toolbar.autosavesConfiguration = (config & ToolbarConfigurationAutoSaveConfiguration) != 0;

    // 设置显示模式
    if (config & ToolbarConfigurationDisplayModeIconOnly) {
        toolbar.displayMode = NSToolbarDisplayModeIconOnly;
    } else if (config & ToolbarConfigurationDisplayModeIconAndText) {
        toolbar.displayMode = NSToolbarDisplayModeIconAndLabel;
    } else if (config & ToolbarConfigurationDisplayModeTextOnly) {
        toolbar.displayMode = NSToolbarDisplayModeLabelOnly;
    }

    window.toolbar = toolbar;

    // 保留委托对象
    objc_setAssociatedObject(window, "MainToolbarDelegate", toolbarDelegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}

#pragma mark - 动态控件创建函数

void AddToolbarButton(unsigned long nsWindowHandle, const char *identifier, const char *title, const char *tooltip, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *titleStr = [NSString stringWithUTF8String:title];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;

    // 创建按钮
    NSButton *button = [NSButton buttonWithTitle:titleStr target:delegate action:@selector(buttonClicked:)];
    button.bezelStyle = style.bezelStyle;
    button.controlSize = style.controlSize;
    if (tooltipStr) {
        button.toolTip = tooltipStr;
    }
    if (style.font) {
        button.font = style.font;
    }

    // 设置尺寸约束
    if (style.width > 0) {
        [button.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [button.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:button forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarImageButton(unsigned long nsWindowHandle, const char *identifier, const char *imageName, const char *tooltip, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *imageNameStr = [NSString stringWithUTF8String:imageName];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;

    // 创建图片按钮
    NSButton *button = [NSButton buttonWithImage:[NSImage imageNamed:imageNameStr]
                                         target:delegate
                                         action:@selector(buttonClicked:)];
    button.bezelStyle = style.bezelStyle;
    button.controlSize = style.controlSize;
    button.imagePosition = NSImageOnly;
    if (tooltipStr) {
        button.toolTip = tooltipStr;
    }
    if (style.font) {
        button.font = style.font;
    }

    // 设置尺寸约束
    if (style.width > 0) {
        [button.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [button.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(button, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:button forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *placeholderStr = placeholder ? [NSString stringWithUTF8String:placeholder] : nil;

    // 创建文本框
    NSTextField *textField = [[NSTextField alloc] init];
    textField.placeholderString = placeholderStr;
    textField.delegate = delegate;
    textField.controlSize = style.controlSize;
    if (style.font) {
        textField.font = style.font;
    }

    // 设置尺寸约束
    if (style.width > 0) {
        [textField.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [textField.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(textField, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:textField forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarSearchField(unsigned long nsWindowHandle, const char *identifier, const char *placeholder, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *placeholderStr = placeholder ? [NSString stringWithUTF8String:placeholder] : nil;

    // 创建搜索框
    NSSearchField *searchField = [[NSSearchField alloc] init];
    searchField.placeholderString = placeholderStr;
    searchField.delegate = delegate;
    searchField.controlSize = style.controlSize;
    if (style.font) {
        searchField.font = style.font;
    }

    // 设置尺寸约束
    if (style.width > 0) {
        [searchField.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [searchField.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(searchField, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:searchField forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarCombobox(unsigned long nsWindowHandle, const char *identifier, const char **items, int count, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 创建下拉框
    NSComboBox *comboBox = [[NSComboBox alloc] init];
    comboBox.delegate = delegate;
    comboBox.controlSize = style.controlSize;
    if (style.font) {
        comboBox.font = style.font;
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
    if (style.width > 0) {
        [comboBox.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [comboBox.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(comboBox, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:comboBox forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

void AddToolbarCustomView(unsigned long nsWindowHandle, const char *identifier, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    NSString *idStr = [NSString stringWithUTF8String:identifier];

    // 创建自定义容器
    NSView *container = [[NSView alloc] init];

    // 设置尺寸约束
    if (style.width > 0) {
        [container.widthAnchor constraintEqualToConstant:style.width].active = YES;
    }
    if (style.height > 0) {
        [container.heightAnchor constraintEqualToConstant:style.height].active = YES;
    }

    // 关联标识符
    objc_setAssociatedObject(container, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    // 添加到委托
    [delegate addControl:container forIdentifier:idStr withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
}

#pragma mark - 工具栏管理函数

void RemoveToolbarItem(unsigned long nsWindowHandle, const char *identifier) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

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

void UpdateToolbarItemStyle(unsigned long nsWindowHandle, const char *identifier, ControlStyle style) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    [delegate updateControlStyle:idStr withStyle:style];
}

void InsertToolbarItemAtIndex(unsigned long nsWindowHandle, const char *identifier, int index) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

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
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return;

    NSString *flexSpaceId = NSToolbarFlexibleSpaceItemIdentifier;

    // 添加到委托（使用nil控件，因为这是系统项）
    ControlStyle style = CreateDefaultControlStyle();
    [delegate addControl:nil forIdentifier:flexSpaceId withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:flexSpaceId atIndex:window.toolbar.items.count];
}

void AddToolbarSpace(unsigned long nsWindowHandle) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return;

    NSString *spaceId = NSToolbarSpaceItemIdentifier;

    // 添加到委托（使用nil控件，因为这是系统项）
    ControlStyle style = CreateDefaultControlStyle();
    [delegate addControl:nil forIdentifier:spaceId withStyle:style];

    // 添加到工具栏
    [window.toolbar insertItemWithItemIdentifier:spaceId atIndex:window.toolbar.items.count];
}

#pragma mark - 控件管理函数

const char *GetToolbarControlValue(unsigned long nsWindowHandle, const char *identifier) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return NULL;

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSView *control = [delegate controlForIdentifier:idStr];

    if (!control) return NULL;

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
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSView *control = [delegate controlForIdentifier:idStr];
    NSString *valueStr = [NSString stringWithUTF8String:value];

    if (!control) return;

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
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");

    if (!delegate) return;

    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSView *control = [delegate controlForIdentifier:idStr];

    if (!control) return;

    if ([control isKindOfClass:[NSControl class]]) {
        [(NSControl *)control setEnabled:(BOOL)enabled];
    }
}