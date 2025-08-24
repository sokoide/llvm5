// StaticLang Loop Examples
// Demonstrates for loops, while loops, and variable scoping

int main() {
    // For loop example
    for (var i int = 0; i < 5; i = i + 1;) {
        print(i);
    }
    
    // While loop example
    var count int = 10;
    while (count > 0) {
        print(count);
        count = count - 1;
    }
    
    // Nested scope example
    var outer int = 100;
    if (outer > 50) {
        var inner int = outer + 50;
        print(inner);
        
        for (var j int = 0; j < 3; j = j + 1;) {
            var nested int = inner + j;
            print(nested);
        }
    }
    
    return 0;
}