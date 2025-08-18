#import "window_config_darwin.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// 工具栏项标识符
const char *BackItemID = "MainIDE.Back";
const char *ForwardItemID = "MainIDE.Forward";
const char *SearchItemID = "MainIDE.Search";
const char *CommandItemID = "MainIDE.Command";

// 全局回调函数指针
static ItemClickCallback gItemClickCallback = NULL;
static SearchTextChangedCallback gSearchTextChangedCallback = NULL;

// 设置回调函数
void SetItemClickCallback(ItemClickCallback callback) {
    gItemClickCallback = callback;
}

void SetSearchTextChangedCallback(SearchTextChangedCallback callback) {
    gSearchTextChangedCallback = callback;
}


// 工具栏委托类
@interface MainToolbarDelegate : NSObject <NSToolbarDelegate>{
    NSMutableDictionary<NSString *, NSTextField *> *_textFields;
}
@property (assign) NSToolbar *toolbar;

@end

@implementation MainToolbarDelegate

- (instancetype)init {
    self = [super init];
    if (self) {
        _textFields = [NSMutableDictionary dictionary];
    }
    return self;
}

- (NSArray<NSToolbarItemIdentifier> *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar {
    return @[
        @(BackItemID),
        @(ForwardItemID),
        @(SearchItemID),
        @(CommandItemID),
        @"TextField1"   // 普通文本框1
    ];
}

- (NSArray<NSToolbarItemIdentifier> *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar {
    return @[
        @(BackItemID),
        @(ForwardItemID),
        @(SearchItemID),
        @(CommandItemID),
        @"TextField1"   // 普通文本框1
    ];
}

- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar
     itemForItemIdentifier:(NSToolbarItemIdentifier)itemIdentifier
 willBeInsertedIntoToolbar:(BOOL)flag {

    NSString *identifier = itemIdentifier;

    if ([identifier isEqualToString:@(BackItemID)]) {
        NSButton *button = [NSButton buttonWithImage:[NSImage imageNamed:NSImageNameGoBackTemplate]
                                                      target:self
                                                      action:@selector(toolbarItemClicked:)];
//         [button setFrameSize:NSMakeSize(48, 48)];              // 大尺寸
        button.bezelStyle = NSBezelStyleTexturedRounded;        // 基础样式
//         button.imageScaling = NSImageScaleProportionallyUpOrDown; // 自适应缩放
        button.identifier = identifier;

        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.view = button;
        item.label = @"后退";
        item.enabled = YES; // 显式启用
//         item.maxSize = NSMakeSize(48, 48);  // 重要：设置最大尺寸
//         item.minSize = NSMakeSize(48, 48);  // 固定尺寸
        // 添加点击事件
        button.target = self;
        button.action = @selector(toolbarItemClicked:);
        objc_setAssociatedObject(button, "itemID", @(BackItemID), OBJC_ASSOCIATION_RETAIN);

        return item;
    }
    else if ([identifier isEqualToString:@(ForwardItemID)]) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.image = [NSImage imageNamed:NSImageNameGoForwardTemplate];
        item.label = @"前进";

        // 添加点击事件
        item.target = self;
        item.action = @selector(toolbarItemClicked:);
        objc_setAssociatedObject(item, "itemID", @(ForwardItemID), OBJC_ASSOCIATION_RETAIN);

        return item;
    }
    else if ([identifier isEqualToString:@(SearchItemID)]) {
        // 创建搜索框
        NSSearchField *searchField = [[NSSearchField alloc] init];
        searchField.placeholderString = @"搜索...";
        [searchField.widthAnchor constraintEqualToConstant:200].active = YES;

        // 添加文本变化事件
        searchField.delegate = self;

        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.view = searchField;
        item.label = @"搜索";
        return item;
    }
    else if ([identifier isEqualToString:@(CommandItemID)]) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.image = [NSImage imageNamed:NSImageNameActionTemplate];
        item.label = @"命令";

        // 添加点击事件
        item.target = self;
        item.action = @selector(toolbarItemClicked:);
        objc_setAssociatedObject(item, "itemID", @(CommandItemID), OBJC_ASSOCIATION_RETAIN);

        return item;
    }else if ([identifier isEqualToString:@"TextField1"]) {
    // 创建普通文本框
        NSTextField *textField = [[NSTextField alloc] init];
        textField.placeholderString = @"输入文本...";
        [textField.widthAnchor constraintEqualToConstant:120].active = YES;

        // 注册文本框
        //_textFields[@"text1"] = textField;

        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.view = textField;
        item.label = @"文本1";
        return item;
    }


    return nil;
}


// 工具栏项点击事件处理
- (void)toolbarItemClicked:(id)sender {
    if (gItemClickCallback) {
        // 获取关联的 itemID
        NSString *itemID = objc_getAssociatedObject(sender, "itemID");
        if (itemID) {
            gItemClickCallback([itemID UTF8String]);
        }
    }
}

// 搜索框文本变化事件处理
- (void)controlTextDidChange:(NSNotification *)notification {
    if (gSearchTextChangedCallback) {
        NSTextField *textField = notification.object;
        if ([textField isKindOfClass:[NSSearchField class]]) {
            gSearchTextChangedCallback([[textField stringValue] UTF8String]);
        }
    }
}

// 搜索框回车事件处理
- (void)controlTextDidEndEditing:(NSNotification *)notification {
    if (gSearchTextChangedCallback) {
        NSTextField *textField = notification.object;
        if ([textField isKindOfClass:[NSSearchField class]]) {
            gSearchTextChangedCallback([[textField stringValue] UTF8String]);
        }
    }
}

@end


// 获取搜索框文本
const char *GetSearchFieldText(unsigned long nsWindowHandle, const char *textName) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    NSToolbar *toolbar = window.toolbar;
    NSString *nsTextName = [NSString stringWithUTF8String:textName];

    for (NSToolbarItem *item in toolbar.visibleItems) {
        if ([item.itemIdentifier isEqualToString:nsTextName]) {
            if ([item.view isKindOfClass:[NSSearchField class]]) {
                NSSearchField *searchField = (NSSearchField *)item.view;
                NSString *text = [searchField stringValue];
                return [text UTF8String];
            }
        }
    }

    return NULL;
}
// 设置搜索框文本
void SetSearchFieldText(unsigned long nsWindowHandle, const char *textName, const char *textValue) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
    NSString *nsTextName = [NSString stringWithUTF8String:textName];
    for (NSToolbarItem *item in window.toolbar.visibleItems) {
        if ([item.itemIdentifier isEqualToString:nsTextName]) {
            if ([item.view isKindOfClass:[NSSearchField class]]) {
                NSSearchField *searchField = (NSSearchField *)item.view;
                [searchField setStringValue:[NSString stringWithUTF8String:textValue]];
                break;
            }
        }
    }
}

// void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *fieldName, const char *placeholder, int width) {
//     NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;
//     MainToolbarDelegate *delegate = objc_getAssociatedObject(window, "MainToolbarDelegate");
//
//     if (!delegate) {
//         NSLog(@"未找到工具栏委托");
//         return;
//     }
//
//     NSString *idStr = [NSString stringWithUTF8String:identifier];
//     NSString *nameStr = [NSString stringWithUTF8String:fieldName];
//     NSString *placeholderStr = [NSString stringWithUTF8String:placeholder];
//
//     // 创建文本框
//     NSTextField *textField = [[NSTextField alloc] init];
//     textField.placeholderString = placeholderStr;
//     [textField.widthAnchor constraintEqualToConstant:width].active = YES;
//
//     // 添加到委托的字典
//     delegate->_textFields[nameStr] = textField;
//
//     // 创建工具栏项
//     NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:idStr];
//     item.view = textField;
//     item.label = placeholderStr;
//
//     // 添加到工具栏
//     [window.toolbar insertItemWithItemIdentifier:idStr atIndex:window.toolbar.items.count];
// }

// 配置窗口
void ConfigureWindow(
    unsigned long nsWindowHandle,
    bool transparentTitleBar,
    int titleBarSeparatorStyle,
    int toolbarStyle,
    int toolbarDisplayMode,
    bool allowsUserCustomization,
    bool autosavesConfiguration
) {
    NSWindow *window = (__bridge NSWindow *)(void *)nsWindowHandle;

    // 配置标题栏
    window.titlebarAppearsTransparent = (BOOL)transparentTitleBar;

    if (@available(macOS 11.0, *)) {
        window.titlebarSeparatorStyle = (NSTitlebarSeparatorStyle)titleBarSeparatorStyle;
    }

    // 创建工具栏
    MainToolbarDelegate *toolbarDelegate = [[MainToolbarDelegate alloc] init];
    NSToolbar *toolbar = [[NSToolbar alloc] initWithIdentifier:@"MainIDE.ToolBar"];
    toolbarDelegate.toolbar = toolbar;
    toolbar.delegate = toolbarDelegate;

    // 配置工具栏样式
    if (@available(macOS 11.0, *)) {
        window.toolbarStyle = (NSWindowToolbarStyle)toolbarStyle;
    }

    toolbar.displayMode = (NSToolbarDisplayMode)toolbarDisplayMode;
    toolbar.allowsUserCustomization = (BOOL)allowsUserCustomization;
    toolbar.autosavesConfiguration = (BOOL)autosavesConfiguration;

    // 设置工具栏
    window.toolbar = toolbar;

    // 保留委托对象
    objc_setAssociatedObject(window, "MainToolbarDelegate", toolbarDelegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}