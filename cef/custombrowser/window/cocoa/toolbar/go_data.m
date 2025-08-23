#import "go_data.h"

// 将GoData转换为Objective-C对象
//id convertGoDataToOC(GoData *data) {
//    if(!data) return nil;
//    switch (data.Type) {
//        case DataType_String: {
//            NSString *str = [NSString stringWithUTF8String:data.DtString];
//            return str;
//        }
//        case DataType_StringArray: {
//            NSMutableArray *array = [NSMutableArray arrayWithCapacity:data.DtStringArray.Count];
//            for (int i = 0; i < data.DtStringArray.Count; i++) {
//                char *cStr = data.DtStringArray.Items[i];
//                NSString *ocStr = [NSString stringWithUTF8String:cStr];
//                [array addObject:ocStr];
//            }
//            return array;
//        }
//        case DataType_Pointer: {
//            // 将void*转换为Objective-C对象
//            return (__bridge id)data.DtPointer;
//        }
//        default:
//            return nil;
//    }
//}

GoData *StringToGo(NSString *string) {
    if (!string) return nil;
    GoData *result = malloc(sizeof(GoData)); // 堆分配GoData
    result->Type = DataType_String;
    result->DtString = (char *)[string UTF8String];
    return result;
}

NSString *StringToOC(GoData *data) {
    if (!data) return nil;
    if (data->Type == DataType_String) {
        NSString *result = [NSString stringWithUTF8String:data->DtString];
        return result;
    }
    return nil;
}

GoData *StringArrayToGo(NSArray<NSString *> *array) {
    if (!array || array.count == 0) return nil;
    int count = (int)array.count;
    char **cArray = (char **)malloc(sizeof(char *) * count);
    for (int i = 0; i < count; i++) {
        NSString *str = array[i];
        cArray[i] = (char *)[str UTF8String];
    }
    GoData *result = malloc(sizeof(GoData)); // 堆分配GoData
    result->Type = DataType_StringArray;
    result->DtStringArray.Items = cArray;
    result->DtStringArray.Count = count;
    return result;
}

NSArray<NSString *> *StringArrayToOC(GoData *data) {
    if (!data) return nil;
    if (data->Type == DataType_String) {
        NSMutableArray *result = [NSMutableArray arrayWithCapacity:data->DtStringArray.Count];
        for (int i = 0; i < data->DtStringArray.Count; i++) {
            char *cStr = data->DtStringArray.Items[i];
            NSString *ocStr = [NSString stringWithUTF8String:cStr];
            [result addObject:ocStr];
        }
        return result;
    }
    return nil;
}

GoData *PointerToGo(id pointer) {
    if (!pointer) return nil;
    GoData *result = malloc(sizeof(GoData)); // 堆分配GoData
    result->Type = DataType_Pointer;
    result->DtPointer = (__bridge void *)pointer;
    return result;
}

id PointerToOC(GoData *data) {
    if (!data) return nil;
    if (data->Type == DataType_Pointer) {
        return (__bridge id)data->DtPointer;
    }
    return nil;
}

void OCFreeGoData(GoData *data) {
    if (!data) return;
    if (data->Type == DataType_String){
        free(data->DtString);
    }
    else if (data->Type == DataType_StringArray) {
        for (int i = 0; i < data->DtStringArray.Count; i++) {
            free(data->DtStringArray.Items[i]); // 释放每个字符串
        }
        free(data->DtStringArray.Items); // 释放指针数组
    }
    else if (data->Type == DataType_Pointer) {
		data->DtPointer = nil;
    }
    free(data); // 释放GoData自身
}