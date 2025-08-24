// StaticLang Fibonacci Example
// Demonstrates functions, recursion, and control flow

function fibonacci(int n) -> int {
    if (n <= 1) {
        return n;
    } else {
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}

int main() {
    int num = 10;
    int result = fibonacci(num);
    print(result);
    return 0;
}