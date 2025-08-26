#import "config.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

void* NewButton(void* nsDelegate, const char *title, const char *tooltip, ControlProperty property) {
    if (!nsDelegate || !title) {
        NSLog(@"[ERROR] NewButton 必要参数为空");
        return nil;
    }
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *titleStr = [NSString stringWithUTF8String:title];
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    NSButton *button = [NSButton buttonWithTitle:titleStr target:delegate action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    ConfigureControl(button, tooltipStr, property);
    return (__bridge void*)(button);
}

void* NewImageButton(void* nsDelegate, NSImage *buttonImage, const char *tooltip, ControlProperty property) {
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    NSButton *button = [NSButton buttonWithImage:buttonImage
                                          target:delegate
                                          action:@selector(buttonClicked:)];
    button.bezelStyle = property.bezelStyle;
    button.imagePosition = NSImageOnly;
    ConfigureControl(button, tooltipStr, property);
    return button;
}

void* NewImageButtonFormImage(void* nsDelegate, const char *image, const char *tooltip, ControlProperty property) {
    NSString *imageNameStr = [NSString stringWithUTF8String:image];
    NSImage *buttonImage = nil;
    // 首先尝试从文件路径加载图像
    NSURL *imageURL = [NSURL fileURLWithPath:imageNameStr];
    if (imageURL) {
        buttonImage = [[NSImage alloc] initWithContentsOfURL:imageURL];
    }
    // 如果文件加载失败，尝试使用系统符号
    if (!buttonImage) {
        buttonImage = [NSImage imageWithSystemSymbolName:imageNameStr accessibilityDescription:nil];
    }
    // 如果仍然没有图像，使用默认图像
    if (!buttonImage) {
        buttonImage = [NSImage imageNamed:NSImageNameActionTemplate];
    }
    return NewImageButton(nsDelegate, buttonImage, tooltip, property);
}

void* NewImageButtonFormBytes(void* nsDelegate, const uint8_t* data, size_t length, const char *tooltip, ControlProperty property) {
    NSImage *buttonImage = imageFromBytes(data, length);
    return NewImageButton(nsDelegate, buttonImage, tooltip, property);
}