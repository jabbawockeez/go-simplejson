### go-simplejson

a Go package to interact with arbitrary JSON

### Importing

    import github.com/jabbawockeez/go-simplejson

### Usage

#### FromString
```
	obj := json.FromString(`{"a":1}`)

	obj.GetPath("a").Append(3)
	obj.Get("a").Append(2)
	obj.GetPath("a").Insert(2, 5)
	// obj.EnSet("a", "b", 6)
	obj.EnSet("b", "bb")
	obj.EnSet("c", "d", "dd")
	// obj.EnSet( 96)
	obj.P()
	s := struct{
		D string
	}{}
	obj.GetPath("c").ToStruct(&s)
	fmt.Printf("%#v", s)
	d3 := obj.Clone()
	fmt.Printf("%#v", obj.Items())
	fmt.Printf("%p %p", d2, d3)
```