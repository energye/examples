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

// 声明回调函数类型
typedef void (*ItemClickCallback)(const char *itemID);
typedef void (*SearchTextChangedCallback)(const char *text);

// 设置回调函数
void SetItemClickCallback(ItemClickCallback callback);
void SetSearchTextChangedCallback(SearchTextChangedCallback callback);
// 获取搜索框文本
const char *GetSearchFieldText(unsigned long nsWindowHandle, const char *textName);
// 设置搜索框文本
void SetSearchFieldText(unsigned long nsWindowHandle, const char *textName, const char *textValue);


// 动态添加文本框
void AddToolbarTextField(unsigned long nsWindowHandle, const char *identifier, const char *fieldName, const char *placeholder, int width);

#ifdef __cplusplus
}
#endif