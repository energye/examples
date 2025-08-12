@echo off

setlocal enabledelayedexpansion

:: 转换所有 .o 文件为 ELF 格式
for %%f in (%*) do (
    if "%%~xf"==".o" (
        zig.exe objcopy coff -O elf64-x86-64 "%%f" "%%f.elf"
        set "args=!args! "%%f.elf""
    ) else (
        set "args=!args! "%%f""
    )
)

zig cc -target x86_64-linux-musl -static %*