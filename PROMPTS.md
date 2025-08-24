Define a small computer programming language (e.g. subset of javascript) including variable/function types. It must be statically typed language.
It supports int, double and string. Variable scope is implemented including global variables.
Garbage collection or automtic reference counting is not necessary. Static memory allocation is fine. If needed, implement `malloc`/`free` like function for heap allocation.
Make a language compiler to compile it into a native binary using llvm/clang. Make the grammer in .y format, compile it with Goyacc.
The compiler is purely written in Go. Make some sample scripts which can be compiled with it.
Make unit test to cover all the written functions as much as possible.
Use program-manager-spec-writer to define the spec, software-architect for the architecture, spec-based-developer for coding, code-tester for testing and code-reviewer for the review.
