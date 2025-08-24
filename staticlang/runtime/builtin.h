/*
 * StaticLang Runtime Built-in Functions Header
 * Provides declarations for memory management and I/O functions
 */

#ifndef STATICLANG_BUILTIN_H
#define STATICLANG_BUILTIN_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Memory management functions */
void* sl_malloc(size_t size);
void sl_free(void* ptr);

/* Print functions for different types */
void sl_print_int(int value);
void sl_print_double(double value);
void sl_print_string(const char* value);

/* String manipulation functions */
char* sl_alloc_string(const char* str);
char* sl_concat_string(const char* str1, const char* str2);
int sl_compare_string(const char* str1, const char* str2);

/* Array allocation */
void* sl_alloc_array(size_t element_size, size_t count);

/* Debug memory functions (only in debug builds) */
#ifdef DEBUG_MEMORY
void* sl_debug_malloc(size_t size, const char* file, int line);
void sl_debug_free(void* ptr, const char* file, int line);
void sl_print_memory_stats(void);
#endif

#ifdef __cplusplus
}
#endif

#endif /* STATICLANG_BUILTIN_H */