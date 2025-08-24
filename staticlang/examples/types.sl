// StaticLang Type System Example
// Demonstrates all supported types and operations

int globalCounter = 0;

function testTypes() -> int {
    int x = 42;
    double pi = 3.14159;
    string name = "StaticLang";
    
    print(x);
    print(pi);
    print(name);
    
    // Arithmetic operations
    int sum = x + 10;
    double product = pi * 2.0;
    
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

int main() {
    int result = testTypes();
    
    // Global variable usage
    globalCounter = result;
    print(globalCounter);
    
    return 0;
}