// StaticLang Loop Examples
// Demonstrates for loops, while loops, and variable scoping

int main() {
    // For loop example
    for (int i = 0; i < 5; i++) {
        print(i);
    }
    
    // While loop example
    int count = 10;
    while (count > 0) {
        print(count);
        count = count - 1;
    }
    
    // Nested scope example
    int outer = 100;
    if (outer > 50) {
        int inner = outer + 50;
        print(inner);
        
        for (int j = 0; j < 3; j++) {
            int nested = inner + j;
            print(nested);
        }
    }
    
    return 0;
}