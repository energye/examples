#import "window_config_darwin.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// 工具栏项标识符
const char *BackItemID = "MainIDE.Back";
const char *ForwardItemID = "MainIDE.Forward";
const char *SearchItemID = "MainIDE.Search";
const char *CommandItemID = "MainIDE.Command";

// 工具栏委托类
@interface MainToolbarDelegate : NSObject <NSToolbarDelegate>
@property (assign) NSToolbar *toolbar;
@end

@implementation MainToolbarDelegate

- (NSArray<NSToolbarItemIdentifier> *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar {
    return @[
        @(BackItemID),
        @(ForwardItemID),
        @(SearchItemID),
        @(CommandItemID)
    ];
}

- (NSArray<NSToolbarItemIdentifier> *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar {
    return @[
        @(BackItemID),
        @(ForwardItemID),
        @(SearchItemID),
        @(CommandItemID)
    ];
}

- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar
     itemForItemIdentifier:(NSToolbarItemIdentifier)itemIdentifier
 willBeInsertedIntoToolbar:(BOOL)flag {

    NSString *identifier = itemIdentifier;

    if ([identifier isEqualToString:@(BackItemID)]) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.image = [NSImage imageNamed:NSImageNameGoBackTemplate];
        item.label = @"后退";
        return item;
    }
    else if ([identifier isEqualToString:@(ForwardItemID)]) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.image = [NSImage imageNamed:NSImageNameGoForwardTemplate];
        item.label = @"前进";
        return item;
    }
    else if ([identifier isEqualToString:@(SearchItemID)]) {
        // 创建搜索框
        NSSearchField *searchField = [[NSSearchField alloc] init];
        searchField.placeholderString = @"搜索...";
        [searchField.widthAnchor constraintEqualToConstant:200].active = YES;

        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.view = searchField;
        item.label = @"搜索";
        return item;
    }
    else if ([identifier isEqualToString:@(CommandItemID)]) {
        NSToolbarItem *item = [[NSToolbarItem alloc] initWithItemIdentifier:identifier];
        item.image = [NSImage imageNamed:NSImageNameActionTemplate];
        item.label = @"命令";
        return item;
    }

    return nil;
}

@end


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