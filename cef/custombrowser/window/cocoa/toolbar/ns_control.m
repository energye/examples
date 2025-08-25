#import "config.h"
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// 创建默认控件样式
ControlProperty CreateDefaultControlProperty() {
    ControlProperty property;
    property.width = 0; // 0表示自动大小
    property.height = 0;
    property.minWidth = 0;
    property.maxWidth = 0;
    property.bezelStyle = NSBezelStyleTexturedRounded;
    property.controlSize = NSControlSizeRegular;
    property.font = nil;
    property.VisibilityPriority = NSToolbarItemVisibilityPriorityStandard;
    return property;
}

// 创建自定义控件样式
ControlProperty CreateControlProperty(CGFloat width, CGFloat height, NSBezelStyle bezelStyle, NSControlSize controlSize, void *font) {
    ControlProperty property;
    property.width = width;
    property.height = height;
    property.bezelStyle = bezelStyle;
    property.controlSize = controlSize;
    property.font = (__bridge NSFont *)font;
    return property;
}

// 通用函数：通过NSControl设置控件属性（适用于按钮、文本框等）
void ConfigureControl(NSControl *control, NSString *tooltipStr, ControlProperty property) {
    control.controlSize = property.controlSize;
    if (tooltipStr) {
        control.toolTip = tooltipStr;
    }
    if (property.font) {
        control.font = property.font;
    }
    if (property.width > 0) {
        [control.widthAnchor constraintEqualToConstant:property.width].active = YES;
    }
    if (property.height > 0) {
        [control.heightAnchor constraintEqualToConstant:property.height].active = YES;
    }
    // 最小和最大宽度约束
    if (property.minWidth > 0) {
        NSLayoutConstraint *minWidthConstraint = [control.widthAnchor constraintGreaterThanOrEqualToConstant:property.minWidth];
        minWidthConstraint.priority = NSLayoutPriorityDefaultHigh;
        minWidthConstraint.active = YES;
    }
    if (property.maxWidth > 0) {
        NSLayoutConstraint *maxWidthConstraint = [control.widthAnchor constraintLessThanOrEqualToConstant:property.maxWidth];
        maxWidthConstraint.priority = NSLayoutPriorityDefaultHigh;
        maxWidthConstraint.active = YES;
    }
}