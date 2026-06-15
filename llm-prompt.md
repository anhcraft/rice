Rice is a minimal embeddable scripting language for Go applications.

The language file extension is `.rice`.

## Data types

```
list.new().append(
    # Primitives
    123,         # integer, 64-bit signed integer
    -123.456,    # float, 64-bit floating-point number
    1e-9,        # float
    true,        # bool
    "Xin chao!", # string
    null,        # null

    # Composites
    list.new(),  # dynamic-typed list
    set.new(),   # dynamic-typed set
    map.new(),   # dynamic-typed map
    func() {     # function
       return
    }
)
```
- Note: There is no array support, there is no array syntax such as `[...]`; use List as an alternative
- Note: There is no dedicated Character/Rune support
- Note: Unicode is fully supported in String; any operation to String involves Unicode and rune-level (NOT byte as in Go)

## Unicode codepoints
```
# A Unicode literal using \u for a 4-digit hexadecimal value
"\u20AC" # €

# A Unicode literal using \U for an 8-digit hexadecimal value
"\U0001F60E" # 😎
```

## Collections
- Collection type can be iterable using `for..in`
- Indexed collection supports random access with element access syntax, e.g. `list[0]`, `map[key]`, `string[0]`
- Mutable collection supports assigning new value to a specific element
- List of collection types and their traits:
  - `string`: indexed, immutable
  - `list`: indexed, mutable
  - `set`: mutable
  - `map`: indexed, mutable
- Examples:
```
"string"[0] # return the first character ("s" of type String)
list.of(1, 2)[0] # return the first element (1)
map.of("key", "value")["key"] # access the "key" entry

# list iterator returns the element value
for elem in list.of(1, 2, 3) {
    elem
}
# set iterator returns the element value
for elem in set.of(1, 2, 3) {
    elem
}
# map iterator returns an entry of list(key, value)
for entry in map.of("key1", "value1", "key2", "value2") {
    const key = entry[0];
    const value = entry[1];
}
# string iterator returns a String for each character; there is no dedicated Character/Rune type
for ch in "hello world" {
    ch
}
```

## Keywords
```
"var", "const", "if", "else", "func", "for", "in", "continue", "break", "return"
```

## Comments
- Only single-line comment is supported, prefixing `#`

---

# Statement
- Statement does not return any value except increment/decrement, expression (a code block is also an expression)
- Simple Statement includes declaration, increment/decrement and expression

## Declaration
```
const name = "Bob";
var age = 19;
```

## For loop
- There is C-style syntax and a short-form syntax
```
# C-style
for (var i = 0; i < n; i++) {

}

# Short-form (while loop)
for i < n {

}
```
- C-style accepts `init: Optional<SimpleStatement>; cond: Optional<Expression>; post-iteration: Optional<SimpleStatement>`

## For in loop
- For-in loop iterates on a collection value and declare a variable to access each element value
```
for elem in list {}
```

## Controls
```
# Loop controls
break;
continue;

# Function controls
return; # no value
return "OK";
return res;
```

## Increment/Decrement
- They are statement, not expression
```
++i # increases i, then read i
i++ # read i, then increases i
--i # decreases i, then read i
i-- # read i, then decreases i
```

# Expression
- Expression can return value; when it is placed on its own, it is also a statement

## Assignment
- Assignment cannot be chained
- Assignment returns the new value
```
var a = 5;
print(a = 10); # print out "10"
#a = 5 = b # chaining is disallowed
```

## Unary and Binary operators
- Follows typical precedence rules
```
1+2*(3 + a);
a && (b || c) && (d > 10);
!check;
-(num/2);
```
- List of operators:
```
'=', '>', '<', '!', '==', '<=', '>=', '!=', '&&', '||', '++', '--',
'+', '-', '*', '/', '%', '...'
```

### Implicit conversion
- Implicit conversion of binary operators, from the highest precedence:
  - First: If either left or right operand is String, convert both into String
  - Second: If either left or right operand is Bool, convert both into Bool
  - Third: If either left or right operand is Float, convert both into Float
  - Otherwise: If either left or right operand is Int, convert both into Int

### Spread
- Spread operator can only be used within an argument list, used to fill a collection value into the argument list; it could be positioned anywhere (does not necessarily at the end)
```
myFunction(...list, 5, ":", ...set)
```

## Block expression
- Block expression is a list of statement. The last statement value is also the block value
- Block expression is used by many other language constructs such as for loop, for-in, if, etc
- A block creates a new lexical scope allowing shadowing variables declared previously
```
var a = 10;
{
    var a = 5;
    print(a); # 5
}
print(a); # 10
```

## If expression
- If expression returns the value of the picked branch
```
var choiceCode = 
    if choice == "Apple" {0}
    else if choice == "Banana" {1}
    else if choice == "Coconut" {2}
    else if choice == "Durian" {3}
    else if choice == "Elderberry" {4};
```

## Element access
- Element access (brackets `[...]`) could be used on collection values in which their type is an indexed collection
- The element could be of any type, it is important to match the correct type with what the collection expects
```
"Hello"[0] # return "H"
map.of("key", "value")["key"] # return "value"
map.of(0, "value")[0] # return "value"
```

## Functions and Call
### Native/built-in functions:
- Having name
- Static-typing: a parameter can accept a specific type, many types (union) or any
- Support overloading, varargs
- Namespaced functions: relevant functions are grouped in a package/namespace, e.g. `strings` has `strings.toLower(), strings.slice()`
- Type-bound functions: attached-functions to certain type of values
    - Transformation rule of type-bound function to namespaced function: `V.f(...) = f(V, ...)`
```
"RICE".toLower() # rice
# equivalent to
strings.toLower("RICE")
```

### User-defined functions (function literal)
- Anonymous: they could be referred by variable name but such name belongs to the variable, not the function literal itself
- Dynamic-typing
- Support varargs
- Does not support overloading
- Function literal captures the innermost lexical scope of where it was defined. It can see changes to that lexical scope from the definition time to when a call is made 
- Function literal is useful in functional programming

```
# An anonymous function literal
func(){print("Hello World")}();

# Define a variable of function
const a = func(){print("Hello World")};
a(); # refer to the function and call it
```

## Selector
- A selector (dot `.`) is used to access a member of a collection by identifier. The identifier is then turned into string key. It is currently only usable for maps
```
map.of("key", "value").key # return "value"
# equivalent to
map.of("key", "value")["key"]
```
- A selector is the primary way to access type-bounded functions
```
"RICE".toLower()
strings.toLower("RICE")
```

## Functional programming
```
strings.join("\n", ...map.new()
    .put(
        "Fruit 1", "Apple",
        "Fruit 2", "Banana",
        "Fruit 3", "Cherry",
        "Fruit 4", "Durian",
        "Fruit 5", "Elderberry"
    )
    .filter(func(entry){
        typeof(entry[1]) == "String" && entry[1].toLower().include("a")
    })
    .entries()
    .sort(func(a,b){a[1]<b[1]})
    .map(func(entry){entry[0]+": "+entry[1]}))
```

# Standard library
- If the signature has ellipse `...`, that function supports varargs
- For parameter type `any`, it accepts any type of values
- The vertical divisor `|` denotes union types
- Native/built-in functions support overloading

## Namespaced functions
### Map
```
map.put(map Map, entries ...any): puts multiple K-V entries in order and return the given map
map.remove(map Map, keys ...any): removes multiple entries by keys and return the given map
map.keys(map Map): returns a set of keys
map.entries(map Map): returns a list of list(key, value)
map.new()
map.of(entries ...any): creates a new map of multiple K-V entries and return the new map
map.includeKey(map Map, key any): checks if a key exists
map.values(map Map): returns a list of value
map.map(map Map, lambda Func): creates a new map from the mapping function of list(key, value)
map.filter(map Map, lambda Func): creates a new map from the filter function of list(key, value)
```

### Datetime
```
datetime.now(): returns the current Unix timestamp in milliseconds.
```

### Type
```
typeof(value any): gets the type name of "Int, Float, Bool, String, List, Set, Map, Func"
float(value any): converts into float
bool(value any): converts into bool
int(value any): converts into int
string(value any): converts into string
isNumberLike(value any): checks if the type is one of Int, Float, Bool
isNumber(value any): checks if the type is either Int or Float
len(value any): gets the length of collection (string, list, set, map)
```

### Error
```
assert(cond Bool, msg String): if the condition fails, throw an exception with the given message
throw(msg String): throw an exception with the given message
```

### I/O
```
println(args ...any)
printlnf(format String, args ...any)
printf(format String, args...any)
print(args ...any)
```

### Strings
```
strings.join(separator String, values ...any): joins multiple values with the separator, implicitly convert each each into String
strings.toUpper(str String)
strings.toLower(str String)
strings.index(str String, substr String)
strings.substr(str String, offset Int, exclusiveEnd Int)
strings.substr(str String, offset Int)
strings.format(fmt String, args ...any)
strings.trim(str String)
strings.include(str String, substr String)
strings.lastIndex(str String, substr String)
strings.split(str String, separator String)
```

### Math
```
math.sqrt(num Int|Float|Bool)
math.floor(num Int|Float|Bool)
math.ceil(num Int|Float|Bool)
math.pow(base Int|Float|Bool, exp Int|Float|Bool)
math.max(args ...any)
math.min(args ...any)
math.abs(num Int|Float|Bool)
```

### List
```
list.slice(list List, offset Int, exclusiveEnd Int): slices the list and returns the new copy
list.slice(list List, offset Int): slices the list and returns the new copy
list.new()
list.of(items ...any)
list.prepend(list List, items ...any): prepends multiple items in-place, return the given list
list.append(list List, items ...any): appends multiple items in-place, return the given list
list.include(list List, item any)
list.index(list List, item any)
list.lastIndex(list List, item any)
list.sort(list List, lambda Func): sorts the list given the comparator of function (a, b), return `true` if a should come before b
list.map(list List, lambda Func): creates a new list from the mapping function of value
list.reverse(list List): reverses the list in place, returns the given list
list.filter(list List, lambda Func): creates a new list from the filter function of value
list.removeAt(list List, index Int): removes an item in-place, returns the given list
list.removeAll(list List, item any): removes all occurrence of the given item, returns the number of items found and removed
```

### Set
```
set.include(set Set, item any)
set.map(set Set, lambda Func): creates a new set from the mapping function of value
set.filter(set Set, lambda Func): creates a new set from the filter function of value
set.remove(set Set, item any): removes an item in-place, returns the given set
set.new()
set.of(items ...any)
set.add(set Set, items ...any): adds multtiple items in-place, returns the given set
```

### JSON
```
json.encode(val any): convert a Rice value into minified JSON
json.encode(val any, indent String): convert a Rice value into prettified JSON
json.decode(val String): convert the given raw JSON into corresponding Rice values 
```

## Type-bound functions
- These are subset of namespaced functions as long as the first parameter accepts the respective type of value
```
(value of type String).toUpper(String)
(value of type String).toLower(String)
(value of type String).index(String,String)
(value of type String).substr(String,Int,Int)
(value of type String).substr(String,Int)
(value of type String).format(String,...any)
(value of type String).trim(String)
(value of type String).include(String,String)
(value of type String).lastIndex(String,String)
(value of type String).split(String,String)
(value of type String).join(String,...any)
(value of type Int).max(...any)
(value of type Int).min(...any)
(value of type Int).abs(Int|Float|Bool)
(value of type Int).sqrt(Int|Float|Bool)
(value of type Int).floor(Int|Float|Bool)
(value of type Int).ceil(Int|Float|Bool)
(value of type Int).pow(Int|Float|Bool,Int|Float|Bool)
(value of type Float).max(...any)
(value of type Float).min(...any)
(value of type Float).abs(Int|Float|Bool)
(value of type Float).sqrt(Int|Float|Bool)
(value of type Float).floor(Int|Float|Bool)
(value of type Float).ceil(Int|Float|Bool)
(value of type Float).pow(Int|Float|Bool,Int|Float|Bool)
(value of type List).append(List,...any)
(value of type List).include(List,any)
(value of type List).index(List,any)
(value of type List).lastIndex(List,any)
(value of type List).sort(List,Func)
(value of type List).map(List,Func)
(value of type List).reverse(List)
(value of type List).filter(List,Func)
(value of type List).removeAt(List,Int)
(value of type List).removeAll(List,any)
(value of type List).slice(List,Int,Int)
(value of type List).slice(List,Int)
(value of type List).prepend(List,...any)
(value of type Set).map(Set,Func)
(value of type Set).filter(Set,Func)
(value of type Set).remove(Set,any)
(value of type Set).add(Set,...any)
(value of type Set).include(Set,any)
(value of type Map).map(Map,Func)
(value of type Map).filter(Map,Func)
(value of type Map).put(Map,...any)
(value of type Map).remove(Map,...any)
(value of type Map).keys(Map)
(value of type Map).entries(Map)
(value of type Map).include(Map,any)
(value of type Map).values(Map)
```

# Reference
## EBNF
```ebnf
(* --- Top Level --- *)
script = [ ws ] , [ statement_sequence ] , [ ws ] , EOF ;
statement_sequence = statement , { ws , ";" , ws , statement } , [ ws , ";" ] ;

(* --- Statements --- *)
statement =
    for_statement
  | for_in_statement
  | control_statement
  | simple_statement
  ;

(* Note: simple_statement is nested within a complex statement. *)
simple_statement =
    declaration_statement
  | inc_dec_statement
  | expression
  ;

(* --- Statement Definitions --- *)

(* Note: no chaining support on RHS *)
declaration_statement = ( "var" | "const" ) , ws , identifier , ws , "=" , ws , logical_or_expression ;

for_statement = "for" , ws , [ for_clause ] , ws , block_expression ;
for_clause =
    ( "(", [ simple_statement ] , ";" , [ expression ] , ";" , [ simple_statement ], ")" ) (* C-style *)
  | expression (* Condition-only *)
  ;

for_in_statement = "for" , ws , identifier , ws , "in" , ws , expression , ws , block_expression ;

(* Note: standalone inc/dec is not expression; cannot return value *)
inc_dec_statement =
    ( "++" | "--" ) , ws , postfix_expression
  | postfix_expression , ws , ( "++" | "--" )
  ;

control_statement =
    "break"
  | "continue"
  | "return" , [ ws , expression ]
  ;

(* --- Expressions (Ordered by Precedence) --- *)
expression = assignment_expression | logical_or_expression ;

(* Note: no chaining support on RHS *)
assignment_expression = postfix_expression , ws , "=" , ws , logical_or_expression ;

logical_or_expression   = logical_and_expression , { ws , "||" , ws , logical_and_expression } ;
logical_and_expression  = comparison_expression , { ws , "&&" , ws , comparison_expression } ;
comparison_expression   = additive_expression , [ ws , comparison_operator , ws , additive_expression ] ;
additive_expression     = multiplicative_expression , { ws , additive_operator , ws , multiplicative_expression } ;
multiplicative_expression = unary_expression , { ws , multiplicative_operator , ws , unary_expression } ;
unary_expression        = { unary_operator , ws } , postfix_expression ;

(* Note: postfix ops can be chained *)
postfix_expression = primary_expression , { element_access | function_call | selector } ;

(* --- Primary Expressions --- *)
primary_expression =
    literal
  | identifier
  | parenthesized_expression
  | block_expression
  | if_expression
  ;

parenthesized_expression = "(" , ws , expression , ws , ")" ;
block_expression = "{" , [ ws ] , [ statement_sequence ] , [ ws ] , "}" ;
if_expression = "if" , ws , expression , ws , block_expression ,
                { ws , "else" , ws , "if" , ws , expression , ws , block_expression } ,
                [ ws , "else" , ws , block_expression ] ;

(* --- Language Constructs --- *)
identifier = ( letter | "_" ) , { letter | digit | "_" } ;
element_access = "[" , ws , expression , ws , "]" ;
selector = "." , ws , identifier ;

function_call = "(" , ws , [ argument_list ] , ws , ")" ;
argument_list = argument , { ws , "," , ws , argument } ;
argument = [ "..." , ws ] , expression ;

(* --- Literals --- *)
literal = string_lit | number_lit | boolean_lit | null_lit | function_lit;

function_lit = "func" , ws , "(" , ws , [ parameter_list ] , ws , ")" , ws , block_expression ;
parameter_list = normal_parameter_list , [ ws , "," , ws , variadic_parameter ] | variadic_parameter ;
normal_parameter_list = identifier , { ws , "," , ws , identifier } ;
variadic_parameter = identifier , ws , "..." ;

string_lit = '"' , { char_in_string } , '"' ;
number_lit = digits , [ "." , digits ] , [ ("e" | "E") , [ "+" | "-" ] , digits ] ;
boolean_lit = "true" | "false" ;
null_lit = "null" ;

(* --- Lexical Tokens --- *)
comparison_operator = "==" | "!=" | ">=" | "<=" | ">" | "<" ;
additive_operator = "+" | "-" ;
multiplicative_operator = "*" | "/" ;
unary_operator = "!" | "-" ;

char_in_string = ? any character except quote or backslash ? | '\' , ( '"' | '\' ) ;
letter = "a"..."z" | "A"..."Z" ;
digit = "0"..."9" ;
digits = digit , { digit } ;

comment = "#" , ? any character except newline ? ;
ws = ( " " | "\t" | "\n" | "\r" | comment )+ ;
```
