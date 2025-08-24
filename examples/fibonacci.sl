// StaticLang Fibonacci Example
// Demonstrates functions, recursion, and control flow

func fibonacci(n int) -> int {
    if (n <= 1) {
        return n;
    } else {
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}

func main() -> int {
    var num int = 10;
    var result int = fibonacci(num);
    print(result);
    return 0;
}