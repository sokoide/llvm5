; ModuleID = 'staticlang'
target datalayout = "e-m:o-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-apple-macosx10.15.0"

; External function declarations
declare i32 @printf(i8*, ...)
declare i8* @malloc(i64)
declare void @free(i8*)

; StaticLang builtin functions
declare void @sl_print_int(i32)
declare void @sl_print_double(double)
declare void @sl_print_string(i8*)
declare i8* @sl_alloc_string(i8*)
declare i8* @sl_concat_string(i8*, i8*)
declare i32 @sl_compare_string(i8*, i8*)
declare i8* @sl_alloc_array(i64, i64)

