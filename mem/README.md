You need to have installed Docker and Make.
## Examples

See `_examples` folder.

## In memory storage configuring

You must set env vars
```
export HOSTNAME_MEM_SRV=localhost
export PORT_MEM_SRV=8080
```

## MEM

1. package mem contains 3 files:
   mem.go (struct that satisfies bucket interface),
   mem_srv.go (server for memory storage),
   mem_test.go
2. After the line ```bucket.StartServerListening()```  the execution flow stops.

