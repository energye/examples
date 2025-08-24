#import "go_arguments.h"

// 释放 GoArguments 及其所有元素
void FreeGoArguments(GoArguments* data) {
    if (!data) return;
    // 释放所有数据项
    for (int i = 0; i < data->Count; i++) {
        GoArgsItem item = data->Items[i];
        // 根据类型释放内存
        switch (item.Type) {
            case ArgsType_Int:
            case ArgsType_Float:
            case ArgsType_Bool:// 基本类型直接释放
                free(item.Value);
                break;
            case ArgsType_String: // 字符串类型
                free(item.Value);
                break;
            case ArgsType_Object: // Objective-C 对象，释放引用
                [(id)item.Value release]; // MRC 手动 release，避免崩溃
                break;
            case ArgsType_Pointer: // 指针类型，不释放指向的内容
                break;
            default:
                break;
        }
    }
    // 释放数组本身
    if (data->Items) {
        free(data->Items);
    }
    // 释放 GoArguments 结构
    free(data);
}

// 通用添加函数
GoArguments* CreateGoArguments(int count, ...) {
    GoArguments* data = malloc(sizeof(GoArguments));
    data->Count = count;
    data->Items = malloc(sizeof(GoArgsItem) * data->Count);

    va_list args;
    va_start(args, count); // 正确初始化：第二个参数是可变参数前的最后一个命名参数（count）

    // 按参数数量count遍历，确保获取所有参数
    for (int i = 0; i < count; i++) {
        id arg = va_arg(args, id); // 逐个获取参数
        GoArgsItem item;
        // 自动类型推断
        if ([arg isKindOfClass:[NSNumber class]]) {
            NSNumber* number = (NSNumber*)arg;
            const char* objCType = [number objCType];
            if (strcmp(objCType, @encode(int)) == 0 ||
                strcmp(objCType, @encode(long)) == 0 ||
                strcmp(objCType, @encode(NSInteger)) == 0) {
                // 整数类型
                int* value = malloc(sizeof(int));
                *value = [number intValue];
                item.Value = value;
                item.Type = ArgsType_Int;
            }
            else if (strcmp(objCType, @encode(float)) == 0 ||
                    strcmp(objCType, @encode (double))) {
                // 浮点数类型
                float* value = malloc(sizeof(float));
                *value = [number floatValue];
                item.Value = value;
                item.Type = ArgsType_Float;
            }
            else if (strcmp(objCType, @encode(BOOL)) == 0 ||
                     strcmp(objCType, @encode(bool)) == 0) {
                // 布尔类型
                bool* value = malloc(sizeof(bool));
                *value = [number boolValue];
                item.Value = value;
                item.Type = ArgsType_Bool;
            }
        }
        else if ([arg isKindOfClass:[NSString class]]) {
            // 字符串类型
            NSString* string = (NSString*)arg;
            char* value = strdup([string UTF8String]);
            item.Value = value;
            item.Type = ArgsType_String;
        }
        else if ([arg isKindOfClass:[NSValue class]] &&
                 strcmp([(NSValue*)arg objCType], @encode(void*)) == 0) {
            // 指针类型（包装在 NSValue 中）
            void* value;
            [(NSValue*)arg getValue:&value];
            item.Value = value;
            item.Type = ArgsType_Pointer;
        }
        else {
            // 对象类型
            item.Value = (void*)[arg retain];
            item.Type = ArgsType_Object;
        }
        // 添加到数组
        data->Items[i] = item;
    }
    va_end(args);
    return data;
}

// 从 GoArguments 获取数据的通用函数
void* GetFromGoArguments(GoArguments* data, int index, GoArgumentsType expectedType) {
    if (!data || index < 0 || index >= data->Count) return NULL;
    GoArgsItem item = data->Items[index];
    if (item.Type != expectedType) return NULL;
    return item.Value;
}

// 类型特定的便捷函数
int GetIntFromGoArguments(GoArguments* data, int index) {
    int* value = (int*)GetFromGoArguments(data, index, ArgsType_Int);
    return value ? *value : 0;
}

float GetFloatFromGoArguments(GoArguments* data, int index) {
    float* value = (float*)GetFromGoArguments(data, index, ArgsType_Float);
    return value ? *value : 0.0f;
}

bool GetBoolFromGoArguments(GoArguments* data, int index) {
    bool* value = (bool*)GetFromGoArguments(data, index, ArgsType_Bool);
    return value ? *value : false;
}

const char* GetStringFromGoArguments(GoArguments* data, int index) {
    return (const char*)GetFromGoArguments(data, index, ArgsType_String);
}

void* GetObjectFromGoArguments(GoArguments* data, int index) {
    return GetFromGoArguments(data, index, ArgsType_Object);
}

void* GetPointerFromGoArguments(GoArguments* data, int index) {
    return GetFromGoArguments(data, index, ArgsType_Pointer);
}