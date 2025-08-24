#import "config.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

void* NewTextField(void* nsDelegate, const char *identifier, const char *placeholder, const char *tooltip, ControlProperty property) {
    if (!nsDelegate || !identifier) {
        NSLog(@"[ERROR] NewTextField 必要参数为空");
        return nil;
    }
    MainToolbarDelegate *delegate = (MainToolbarDelegate*)nsDelegate;
    NSString *idStr = [NSString stringWithUTF8String:identifier];
    NSString *placeholderStr = placeholder ? [NSString stringWithUTF8String:placeholder] : nil;
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : nil;
    NSTextField *textField = [[NSTextField alloc] init];
    textField.placeholderString = placeholderStr;
    textField.delegate = delegate;

    // 设置自动调整大小的属性
    [textField setContentHuggingPriority:NSLayoutPriorityDefaultLow
                          forOrientation:NSLayoutConstraintOrientationHorizontal];
    [textField setContentCompressionResistancePriority:NSLayoutPriorityDefaultLow
                                        forOrientation:NSLayoutConstraintOrientationHorizontal];
    objc_setAssociatedObject(textField, @"identifier", idStr, OBJC_ASSOCIATION_RETAIN);
    ConfigureControl(textField, tooltipStr, property);
    return (__bridge void*)(textField);
}