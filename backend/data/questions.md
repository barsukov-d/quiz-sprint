  Go (основная часть ~90 вопросов):

  | Вопрос                                 | Правильный ответ                              |
  |----------------------------------------|-----------------------------------------------|
  | What keyword exits a loop early?       | break                                         |
  | What is the empty interface?           | interface{}                                   |
  | What does 'defer' do?                  | Delays execution until function returns       |
  | What does recover() do?                | Catches panic                                 |
  | What command initializes a new module? | go mod init                                   |
  | How are exported identifiers named?    | Uppercase                                     |
  | Dereference nil pointer?               | Panics                                        |
  | NOT a basic type in Go?                | char                                          |
  | Test function signature?               | func TestXxx(t *testing.T)                    |
  | Blank identifier '_'?                  | Ignores a value                               |
  | Variadic functions?                    | Functions accepting variable arguments        |
  | go mod tidy?                           | Removes unused dependencies                   |
  | Which loop keyword in Go?              | for                                           |
  | Test files suffix?                     | _test.go                                      |
  | Method receiver?                       | A special first argument                      |
  | Buffered channel capacity 5?           | make(chan int, 5)                             |
  | Read from nil channel?                 | Yes, but blocks forever                       |
  | new() vs make()?                       | new for value types, make for reference types |
  | Length of a slice?                     | len(s)                                        |
  | go mod vendor?                         | Copies dependencies to vendor/                |
  | Testing package?                       | testing                                       |
  | Tests with coverage?                   | go test -cover                                |
  | Check interface implementation?        | Type assertion                                |
  | select statement?                      | Multiplexing channel operations               |
  | Interface embedding?                   | Including one interface in another            |
  | Pointer vs value receivers?            | Pointer can modify, value cannot              |
  | Run benchmarks?                        | go test -bench                                |
  | Array capacity?                        | Fixed at declaration                          |
  | Constant keyword?                      | const                                         |
  | Infinite loop?                         | for {}                                        |
  | Map in Go?                             | A hash table                                  |
  | Channel in Go?                         | A type for goroutine communication            |
  | Multiple return values?                | Yes                                           |
  | Send to closed channel?                | Panics                                        |
  | Comma-ok idiom?                        | Value and boolean                             |
  | Type assertion syntax?                 | value.(Type)                                  |
  | Benchmark signature?                   | func BenchmarkXxx(b *testing.B)               |
  | Zero value integer?                    | 0                                             |
  | Entry point?                           | main()                                        |
  | t.Error() vs t.Fatal()?                | Fatal stops test immediately                  |
  | When use panic()?                      | For unrecoverable errors                      |
  | Skip test conditionally?               | t.Skip()                                      |
  | Key exists in map?                     | value, exists := map[key]                     |
  | Import for fmt.Println?                | fmt                                           |
  | Error handling keyword?                | defer                                         |
  | Add dependency?                        | go get                                        |
  | Executable package name?               | main                                          |
  | Error type?                            | error interface                               |
  | Idiomatic error handling?              | Check if err != nil                           |
  | Goroutine?                             | A lightweight thread                          |
  | Return pointer to local var?           | Yes, moved to heap                            |
  | new() function?                        | Allocates memory and returns pointer          |
  | Methods on any type?                   | Yes, on any type defined in same package      |
  | b.N in benchmarks?                     | Number of iterations                          |
  | Unsafe pointer package?                | unsafe                                        |
  | Short variable declaration?            | x := 10                                       |
  | Zero value of pointer?                 | nil                                           |
  | Purpose of go.sum?                     | Dependency checksums                          |
  | Manual memory management?              | No, has garbage collector                     |
  | Mark test as failed?                   | t.Fail()                                      |
  | Create simple error?                   | errors.New()                                  |
  | sync package?                          | Synchronization primitives                    |
  | Single-line comment?                   | // comment                                    |
  | <- operator with channels?             | Sends or receives values                      |
  | Close a channel?                       | close(channel)                                |
  | Multiple interfaces?                   | Yes                                           |
  | Short variable declaration symbol?     | :=                                            |
  | go.mod file?                           | Defines a Go module                           |
  | Interfaces nil?                        | Yes                                           |
  | Variable keyword?                      | var                                           |
  | Multiple files in package?             | Yes, same directory                           |
  | Zero value of slice?                   | nil                                           |
  | Address operator?                      | &                                             |
  | Signal goroutine completion?           | Channel or WaitGroup                          |
  | Empty interface methods?               | 0                                             |
  | Blank identifier for unused returns?   | _                                             |
  | Create goroutine keyword?              | go                                            |
  | Dereference operator?                  | *                                             |
  | Import by relative path?               | No, use module path                           |
  | Interface in Go?                       | A set of method signatures                    |
  | Pointer arithmetic?                    | No, not supported                             |
  | NOT a programming paradigm?            | Sequential                                    |

  JavaScript/TypeScript (~20 вопросов):

  | Вопрос                    | Правильный ответ                          |
  |---------------------------|-------------------------------------------|
  | array.map() returns?      | A new array with transformed elements     |
  | async/await purpose?      | Handle async in synchronous-looking way   |
  | Closure in JS?            | Function with access to outer scope       |
  | Object.freeze()?          | Prevent modifications to an object        |
  | Spread operator (...)?    | Expands iterable into individual elements |
  | console.log(typeof null)? | object                                    |
  | let vs const?             | const cannot be reassigned                |
  | Event bubbling?           | Events travel from child to parent        |
  | .then() method?           | Execute code after Promise resolves       |
  | Language of the web?      | JavaScript                                |
  | What does HTML stand for? | Hyper Text Markup Language                |

  Python (~5 вопросов):

  | Вопрос              | Правильный ответ                          |
  |---------------------|-------------------------------------------|
  | "self" keyword?     | The instance of the class                 |
  | "with" statement?   | Ensures proper resource cleanup           |
  | List comprehension? | A concise way to create lists             |
  | Decorator?          | A function that modifies another function |
  | Tuple vs list?      | Tuples immutable, lists mutable           |

  General Knowledge (~25 вопросов):

  | Вопрос                      | Правильный ответ                               |
  |-----------------------------|------------------------------------------------|
  | Capital of France?          | Paris                                          |
  | Capital of Japan?           | Tokyo                                          |
  | Capital of Australia?       | Canberra                                       |
  | Smallest country?           | Vatican City                                   |
  | Largest population?         | India                                          |
  | Most time zones?            | France                                         |
  | Longest river?              | Nile River                                     |
  | Largest desert?             | Antarctic Desert                               |
  | Largest museum?             | The Louvre                                     |
  | Europe/Asia mountain range? | Ural Mountains                                 |
  | Sistine Chapel city?        | Vatican City                                   |
  | Speed of light?             | 299,792,458 m/s                                |
  | Pyramids builders?          | Ancient Egyptians                              |
  | Berlin Wall fell?           | 1989                                           |
  | WWII ended?                 | 1945                                           |
  | Mona Lisa painter?          | Leonardo da Vinci                              |
  | Romeo and Juliet author?    | William Shakespeare                            |
  | Four Seasons composer?      | Antonio Vivaldi                                |
  | Communist Manifesto?        | Karl Marx and Friedrich Engels                 |
  | First US president?         | George Washington                              |
  | First Harry Potter film?    | 2001                                           |
  | Inception director?         | Christopher Nolan                              |
  | Iron Man actor?             | Robert Downey Jr.                              |
  | Best Picture 2020?          | Parasite                                       |
  | Highest-grossing film?      | Avatar                                         |
  | Newton's First Law?         | Object at rest stays at rest unless acted upon |
  | Gravity?                    | Force attracting objects with mass             |
  | Kinetic energy?             | Energy of motion                               |
  | Formula for force?          | F = ma                                         |
  | Binary search complexity?   | O(log n)                                       |
  | Git purpose?                | Track changes in code over time                |