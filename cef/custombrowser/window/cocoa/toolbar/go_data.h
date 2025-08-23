#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

#ifdef __cplusplus
extern "C" {
#endif

// 数据类型枚举
typedef enum {
    DataType_None,
    DataType_String,
    DataType_StringArray,
    DataType_Pointer
} GoDataType;

// Go交互数据 string array
typedef struct {
    char** Items;
    int Count;
} StringArray;

// Go交互数据
typedef struct {
    GoDataType Type;              // 数据类型
    char* DtString;             // 字符串
    StringArray DtStringArray;  // 字符串数组
    void* DtPointer;            // 实例指针
} GoData;

// GoData 数据转换
GoData *StringToGo(NSString *string);
NSString *StringToOC(GoData *data);
GoData *StringArrayToGo(NSArray<NSString *> *array);
NSArray<NSString *> *StringArrayToOC(GoData *data);
GoData *PointerToGo(id pointer);
id PointerToOC(GoData *data);
// GoData 释放
void OCFreeGoData(GoData *data);
void GoFreeGoData(GoData *data);

#ifdef __cplusplus
}
#endif