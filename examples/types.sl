// StaticLang Type System Example
// Demonstrates all supported types and operations

int globalCounter = 0;

func testTypes() -> int {
    var x int = 42;
    var pi float = 3.14159;
    var name string = "StaticLang";

    print(x);
    print(pi);
    print(name);

    // Arithmetic operations
    var sum int = x + 10;
    var product float = pi * 2.0;

    print(sum);
    print(product);

    // Comparison operations
    if (x > 40) {
        print("x is greater than 40");
    }

    if (pi < 4.0) {
        print("pi is less than 4");
    }

    return sum;
}

func main() {
    var result int = testTypes();

    // Global variable usage
    globalCounter = result;
    print(globalCounter);

    return 0;
}