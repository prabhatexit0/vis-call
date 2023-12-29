# Vis Call

Just a fun project to visualize functions in different programming languages.
String with JS first :) 


```js
function A(bool) {
    if (bool) {
        return B()
    } else {
        return C()
    }
    D()
}

function B(bool) {
}

function C(bool) {
}
```

## Visual Output
```
         A [Doc] [Custom Comments]
        / \   \\
       /   \   \\
      /     \   \\
     /       \   D [D][C]
    /         \
  B [D][C]    C [D][C]



     B [Doc] [Custom Comments]
    / \
   /   \
  nil  nil



     C [Doc] [Custom Comments]
    / \
   /   \
  nil  nil



     D [Doc] [Custom Comments]
    / \
   /   \
  nil  nil
```

## Types
```
Function Declaration Node ->> 
    Name [string]
    Probable Call Expressions [List<Call Expressions>]
    Definite Call Expressions [List<Call Expressions>]
    Doc string [Optional<string>]
    * Custom comments -> can be added by users


Call Expression Node ->> 
    Name [string]
    * Custom comments -> can be added by users
    ? some information about call site
```

