# Docile Monkey

Docile Monkey is an extremely simple HTTP server that responds what you want it to respond.

It may be useful to test the resiliency of your application.
For example, you may use it to easily check how your application reacts when it receives unexpected HTTP responses (code 500, whatever).
It could also make it easy to test circuit breakers.

## Usage

### As a standalone server

Download Docile Monkey binary and launch it.

Parameters:
- **listen**: address on which the server will listen for requests. Default value: **:8080**

Example:

```
$ ./docilemonkey -listen=:8080
```

### In your tests (for Go projects)

You can use it in your test through the httptest.Server

```go
func TestSomething(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(docilemonkey.Handler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

    //...
}
```

