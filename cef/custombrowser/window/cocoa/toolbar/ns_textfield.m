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


// 通过指针获取搜索框的值
const char* GetTextFieldText(void* ptr) {
    NSTextField* textField = (__bridge NSTextField*)(ptr);
    NSString* nsText = [textField stringValue];
    // 转换为 C 字符串（需注意：返回的指针需在 Go 中及时处理，避免被释放）
    return [nsText UTF8String];
}

// 通过指针设置搜索框文本
void SetTextFieldText(void* ptr, const char* text) {
    NSTextField* textField = (__bridge NSTextField*)(ptr);
    NSString* nsText = [NSString stringWithUTF8String:text];
    [textField setStringValue:nsText];
}

// 通过指针设置搜索框文本
void UpdateTextFieldWidth(void* ptr, CGFloat width) {
    NSTextField* textField = (__bridge NSTextField*)(ptr);
    // 移除现有宽度约束
    for (NSLayoutConstraint *constraint in textField.constraints) {
        if (constraint.firstAttribute == NSLayoutAttributeWidth) {
            [textField removeConstraint:constraint];
            break;
        }
    }
    // 添加新宽度约束并设置高优先级
    NSLayoutConstraint *widthConstraint = [textField.widthAnchor constraintEqualToConstant:width];
    widthConstraint.priority = NSLayoutPriorityRequired;
    widthConstraint.active = YES;
}
