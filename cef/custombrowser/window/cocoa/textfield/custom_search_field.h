#import <Cocoa/Cocoa.h>

#ifdef __cplusplus
extern "C" {
#endif

// 回调函数类型定义（供C接口使用）
typedef void (*LeftClickCallback)(void* userData);
typedef void (*RightClickCallback)(void* userData);
typedef void (*TextChangedCallback)(const char* text, void* userData);

@interface CustomSearchField : NSTextField

// 初始化方法
- (instancetype)initWithFrame:(NSRect)frame;

// 设置图标
- (void)setLeftImageWithPath:(NSString*)imagePath;
- (void)setRightImageWithPath:(NSString*)imagePath;

// 设置文本对齐
- (void)setCustomAlignment:(NSTextAlignment)alignment;

// 设置回调
- (void)setLeftClickCallback:(LeftClickCallback)callback userData:(void*)userData;
- (void)setRightClickCallback:(RightClickCallback)callback userData:(void*)userData;
- (void)setTestChangedCallback:(TextChangedCallback)callback userData:(void*)userData;

@end

#ifdef __cplusplus
}
#endif
