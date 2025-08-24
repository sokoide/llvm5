// StaticLang Type System Example
// Demonstrates all supported types and operations

func testTypes() -> int {
    var x int = 42;
    var sum int = x + 10;

    // Comparison operations
    if (x > 40) {
        var message int = 1;
    }

    return sum;
}

func main() -> int {
    var result int = testTypes();
    return result;
}