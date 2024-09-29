### go-simplejson

a Go package to interact with arbitrary JSON

### Importing

    import github.com/jabbawockeez/go-simplejson

### Usage

#### Create a json object
```
    obj := json.New()

    obj = json.FromString(`{"a": 1}`)

    // FromString also accept byte array
    obj = json.FromString([]byte{'{', '}'})

    obj, _ = json.FromFile("/path/to/file")
```

#### pretty print

```
    obj := json.FromString(`{"a": 1}`)
    obj.P()

    /*
    output:
    {
        "a": 1
    }
    */

```

#### EnSet
```
    obj.EnSet("b", "c", 2)
    obj.EnSet("d", []int{3,4,5})
    obj.EnSet("e", json.FromString(`{"f":6}`))

    /*
    output:
    {
        "a": 1,
        "b": {
            "c": 2
        },
        "d": [
            3,
            4,
            5
        ],
        "e": {
            "f": 6
        }
    }
    */
```

#### Get、GetPath、GetIndex
```
    obj.Get("a").P()
    obj.GetPath("b", "c").P()
    obj.GetPath("d", 1).P()
    obj.Get("d").GetIndex(2).P()

    /*
    output:
    1
    2
    4
    5
    */
```

#### Length
```
    fmt.Println(obj.Length())
    fmt.Println(obj.Get("d").Length())

    /*
    output:
    4
    3
    */
```


#### Insert、Append
```
    obj.Get("d").Insert(1, "hello")
    obj.Get("d").Append("world")
    obj.Get("d").Extend([]string{"simple", "json"})
    obj.Get("d").Extend(json.FromString(`["so", "easy"]`))

    /*
    output:
    {
        ... 
        "d": [
            3,
            "hello",
            4,
            5,
            "world",
            "simple",
            "json",
            "so",
            "easy"
        ],
        ...
    */
```

#### Del、DelIndex
```
    obj.Del("a")
    obj.Get("b").Del("c")
    obj.Get("d").DelIndex(2)
```
output:
```
{
  "b": {},
  "d": [
    3,
    4
  ]
}
```


#### Keys、Items
```
    fmt.Println(obj.Keys())
    /*
    output:
    [a b d]
    */

    // Items can be used to iterate over dict and array
    for key, value := range obj.Items() {
        fmt.Println(key, value.ToString())
        value.P()
    }
    /*
    output:
    a 1
    b {"c":2}
    d [3,"hello",4,5,"world","json","simple"]
    */

    for key, value := range obj.Get("d").Items() {
        fmt.Println(key, value.ToString())
        value.P()
    }
    /*
    output:
    4 "world"
    5 "simple"
    6 "json"
    0 3
    1 "hello"
    2 4
    3 5
    */
```

#### get value
```
    /*
        data:
        {
            "a": 1,
            "b": "234",
            "c": ["11", "22"],
            "e": {
                "f": 5
            },
            "g": true,
            "h": [
                {"i": 111},
                {"i": 222}
            ]
        }
    */
    obj.GetInt("a")
    obj.GetInt64("a")
    obj.GetFloat64("a")

    obj.GetString("b")

    obj.GetArray("c")
    obj.GetStringArray("c")

    obj.GetInt("e", "f")
    obj.GetMap("e")

    obj.GetBool("g")

    obj.GetInt("h", 1, "i") // output: 222

```

### type conversion
#### string -> struct
```
   s := `{
        "a": 1,
        "b": "234"
    }`

    type T struct {
        A     int
        B     string
    }

    var t T

    obj := json.FromString(s)
    obj.ToStruct(&t)

    // or just use StringToStruct for short
    // json.StringToStruct(s, &t)
```


#### struct -> string
```
   type T struct {
        A     int
        B     string
    }

    var t = T{1, "222"}

    obj := json.FromStruct(t)
    obj.ToString()

    // or just use StructToString for short
    json.StructToString(t)
```


#### Clone
```
   obj2 := obj.Clone()
```