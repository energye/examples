#import <Cocoa/Cocoa.h>

@interface CustomView : NSView

// 添加颜色属性实现动态修改
@property (nonatomic, strong) NSColor *fillColor;

@end
