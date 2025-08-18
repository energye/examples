#import <AppKit/AppKit.h>

#ifdef __cplusplus
extern "C" {
#endif

// 配置窗口
void ConfigureWindow(
    unsigned long nsWindowHandle,
    bool transparentTitleBar,
    int titleBarSeparatorStyle,
    int toolbarStyle,
    int toolbarDisplayMode,
    bool allowsUserCustomization,
    bool autosavesConfiguration
);

// 工具栏项标识符
extern const char *BackItemID;
extern const char *ForwardItemID;
extern const char *SearchItemID;
extern const char *CommandItemID;


#ifdef __cplusplus
}
#endif