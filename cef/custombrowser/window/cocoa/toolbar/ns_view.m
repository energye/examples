#import "config.h"
#import "ns_view.h"

@implementation CustomView

- (instancetype)initWithFrame:(NSRect)frame {
    self = [super initWithFrame:frame];
    if (self) {
        _fillColor = [NSColor blueColor]; // 设置默认颜色
    }
    return self;
}

- (void)drawRect:(NSRect)dirtyRect {
    // 使用当前fillColor属性值填充视图
    [self.fillColor set];
    NSRectFill(dirtyRect);
    [super drawRect:dirtyRect];
}

@end

void* NewCustomView(void* nsDelegate, const char *identifier) {
    if (!nsDelegate || !identifier) {
        NSLog(@"[ERROR] NewTextField 必要参数为空");
        return nil;
    }
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *idStr = [NSString stringWithUTF8String:identifier];

    CustomView *customView = [[CustomView alloc] init];
    customView.fillColor = [NSColor systemBlueColor]; // 设置填充颜色

    NSRect frame = customView.frame;
    frame.size.width = 150;  // 设置宽度
    frame.size.height = 30; // 设置高度
    customView.frame = frame;

    objc_setAssociatedObject(customView, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);

    return (__bridge void*)(customView);
}