/*
 * StaticLang Runtime Built-in Functions
 * Provides memory management and I/O functions for StaticLang programs
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/*
 * Memory allocation function similar to malloc
 * Returns a pointer to allocated memory or NULL on failure
 */
void* sl_malloc(size_t size) {
    return malloc(size);
}

/*
 * Memory deallocation function similar to free
 * Frees memory allocated by sl_malloc
 */
void sl_free(void* ptr) {
    if (ptr != NULL) {
        free(ptr);
    }
}

/*
 * Print function for integers
 * Prints an integer value followed by a newline
 */
void sl_print_int(int value) {
    printf("%d\n", value);
}

/*
 * Print function for doubles
 * Prints a double value followed by a newline
 */
void sl_print_double(double value) {
    printf("%f\n", value);
}

/*
 * Print function for strings
 * Prints a string value followed by a newline
 */
void sl_print_string(const char* value) {
    if (value != NULL) {
        printf("%s\n", value);
    }
}

/*
 * String allocation and initialization
 * Allocates memory for a string and copies the content
 */
char* sl_alloc_string(const char* str) {
    if (str == NULL) return NULL;
    
    size_t len = strlen(str);
    char* result = malloc(len + 1);
    if (result != NULL) {
        strcpy(result, str);
    }
    return result;
}

/*
 * String concatenation
 * Allocates new memory for concatenated string
 */
char* sl_concat_string(const char* str1, const char* str2) {
    if (str1 == NULL && str2 == NULL) return NULL;
    if (str1 == NULL) return sl_alloc_string(str2);
    if (str2 == NULL) return sl_alloc_string(str1);
    
    size_t len1 = strlen(str1);
    size_t len2 = strlen(str2);
    char* result = malloc(len1 + len2 + 1);
    
    if (result != NULL) {
        strcpy(result, str1);
        strcat(result, str2);
    }
    
    return result;
}

/*
 * String comparison
 * Returns 0 if strings are equal, non-zero otherwise
 */
int sl_compare_string(const char* str1, const char* str2) {
    if (str1 == NULL && str2 == NULL) return 0;
    if (str1 == NULL || str2 == NULL) return 1;
    return strcmp(str1, str2);
}

/*
 * Array allocation
 * Allocates memory for an array of the specified type and size
 */
void* sl_alloc_array(size_t element_size, size_t count) {
    return calloc(count, element_size);
}

/*
 * Memory debugging functions (only active in debug builds)
 */
#ifdef DEBUG_MEMORY
static size_t allocated_bytes = 0;
static size_t allocation_count = 0;

void* sl_debug_malloc(size_t size, const char* file, int line) {
    void* ptr = malloc(size);
    if (ptr) {
        allocated_bytes += size;
        allocation_count++;
        fprintf(stderr, "ALLOC: %zu bytes at %p (%s:%d)\n", size, ptr, file, line);
    }
    return ptr;
}

void sl_debug_free(void* ptr, const char* file, int line) {
    if (ptr) {
        allocation_count--;
        fprintf(stderr, "FREE: %p (%s:%d)\n", ptr, file, line);
        free(ptr);
    }
}

void sl_print_memory_stats() {
    fprintf(stderr, "Memory Stats - Allocated: %zu bytes, Active allocations: %zu\n", 
            allocated_bytes, allocation_count);
}

#define sl_malloc(size) sl_debug_malloc(size, __FILE__, __LINE__)
#define sl_free(ptr) sl_debug_free(ptr, __FILE__, __LINE__)
#endif