# flag

this package is used to flag some logic code or route if you don't want to publish on production environment. It's verry usefull if you use TBD (Trunk Based Development) and many people work on the same codebase.

with this package we can enable or disable some logic without redeploy application. so we expect to reduce the risk of deployment.

## Dependencies
its use goimports to format the code, so you need to install goimports first

## Usage

### On router

```go

func main(){
    var db *sql.DB
    var redis *redis.Client

    _flag := flag.Init(db, redis)

    mux := http.NewServeMux()
    server.New(mux, _flag) // this outomaticly add flag to all route from flag

}
```

### On logic code

if u want to flag some logic code, u can use flag like this:

```go

func SomeLogicFunc(){
	if flag.IsFlagged("someFlag"){ 
		// do something with active flag
	}else{ 
		// do something with inactive flag 
	}
}
```

if u want to flag some route, u can use flag like this:

`note: currently we only support http from standard library`

```go
http.HandleFunc("GET /some-route", flag.ReleaseFlag("some-flag", handler))
```

#### How to autoremoved flag ?

1. build flag cli
``` bash 
go build -o flag cmd/flag/main.go
```

2. run flag cli
Auto remove flag after one month
```bash
./flag erase --db-driver="postgres" --dsn="postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable" --root-path="."
```
or specific flag in our code
```bash
./flag erase --flags="some-flag,some-flag-2,some-flag-3" --root-path="." --db-driver="postgres" --dsn="postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable"
```
`--db-driver` is driver of your database [mysql, postgres]

`--dsn` is connection string to your database

`--flags` is flags that you want to remove

`--root-path` is path to your project

    example: ./pkg/flag or "."
    default value is "."


how to work, example file:

```go
func TestCoba123() {

    var segmentation int
    
    if flag.IsEnable("some-flag") {
        fmt.Println("yeah....")
    }else{
        fmt.Println("woooo.......")
    }
    
}
```

after run command with arg `--flags=some-flag` should be
```go
func TestCoba123() {

    var segmentation int


    fmt.Println("yeah....")


    // some code
}
```

or if code have negation
``` go
func TestCoba123() {

    var segmentation int
    
    if !flag.IsEnable("some-flag") {
        fmt.Println("yeah....")
    }else{
        fmt.Println("woooo.......")
    }
    
    // some code
}
```
after run code `--flags=some-flag` should be
```go
func TestCoba123() {

    var segmentation int


    fmt.Println("woooo.......")


    // some code
}
```

or if on route middleware

from

```go
http.HandleFunc("GET /some-route", flag.ReleaseFlag("some-flag", handler))
```
after run should be
``` go 
http.HandleFunc("GET /some-route", handler)
```