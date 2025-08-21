#ifndef SEARCHFIELD_BRIDGE_H
#define SEARCHFIELD_BRIDGE_H

#include <stdint.h>
#include <stdbool.h>

// 确保C++编译器正确处理C函数命名
#ifdef __cplusplus
extern "C" {
#endif

// 前向声明
typedef struct CustomSearchField CustomSearchField;
typedef CustomSearchField* CustomSearchFieldRef;

// 创建搜索框
CustomSearchFieldRef custom_search_field_create(int x, int y, int width, int height);

// 设置图标
void custom_search_field_set_left_image(CustomSearchFieldRef ref, const char* imagePath);
void custom_search_field_set_right_image(CustomSearchFieldRef ref, const char* imagePath);

// 设置文本对齐
void custom_search_field_set_alignment(CustomSearchFieldRef ref, int alignment);

// 文本操作
void custom_search_field_set_text(CustomSearchFieldRef ref, const char* text);
const char* custom_search_field_get_text(CustomSearchFieldRef ref);

// 回调函数类型定义
typedef void (*LeftClickCallback)(void* userData);
typedef void (*RightClickCallback)(void* userData);
typedef void (*TextChangedCallback)(const char* text, void* userData);

// 设置回调
void custom_search_field_set_left_click_callback(
    CustomSearchFieldRef ref,
    LeftClickCallback callback,
    void* userData
);

void custom_search_field_set_right_click_callback(
    CustomSearchFieldRef ref,
    RightClickCallback callback,
    void* userData
);

void custom_search_field_set_text_changed_callback(
    CustomSearchFieldRef ref,
    TextChangedCallback callback,
    void* userData
);

// 获取视图指针
void* custom_search_field_get_view(CustomSearchFieldRef ref);

// 释放资源
void custom_search_field_destroy(CustomSearchFieldRef ref);

#ifdef __cplusplus
}
#endif

#endif /* SEARCHFIELD_BRIDGE_H */
