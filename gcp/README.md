## How to run tests

```
make test
```

You need to have installed Docker and Make.
## Examples

See `_examples` folder.

## GCP configuring

To run the client library, you must first set up authentication. One way to do that is to create a service account and set an environment variable

```
export GOOGLE_APPLICATION_CREDENTIALS="KEY_PATH"
```

## Helpful links

[https://pkg.go.dev/cloud.google.com/go/storage](https://pkg.go.dev/cloud.google.com/go/storage)

[https://cloud.google.com/storage/docs/reference/libraries#client-libraries-install-go](https://cloud.google.com/storage/docs/reference/libraries#client-libraries-install-go)

[https://cloud.google.com/storage/docs/samples/storage-generate-signed-url-v4](https://cloud.google.com/storage/docs/samples/storage-generate-signed-url-v4)

## GCP 

1. package GCP contains 3 files:
   gcp.go (struct that satisfies bucket interface),
   adapter_gcp.go (adapter layer that needed for mock),
   gcp_test.go
2. to run this, you should set env GOOGLE_APPLICATION_CREDENTIALS
   by your JSON format credentials. Or put it to secrets directories 
   as key.json and use MakeFile(or <code>make run</code>).
3. to run tests you should generate mocks by <code>make gen-mocks-gcp</code>.

