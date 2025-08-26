#import <Cocoa/Cocoa.h>

#ifdef __cplusplus
extern "C" {
#endif

@interface CustomView : NSView

@property (strong, nonatomic) NSColor *backgroundColor;

@end

void* NewCustomView(const char *identifier);

#ifdef __cplusplus
}
#endif