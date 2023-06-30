# README

## Development

### Testing

All tests:

```
go test -tags all_tests ./...
```

#### Unit tests

All unit tests:

```
go test -tags all_unit_tests ./...
```

Unit tests per feature:

```
go test -tags unit_tests,packages ./...
go test -tags unit_tests,database ./...
go test -tags unit_tests,repository ./...
```

#### Integration tests

All integration tests:

```
go test -tags all_integration_tests ./...
```

Integration tests per feature:

```
go test -tags integration_tests,packages ./...
go test -tags integration_tests,database ./...
go test -tags integration_tests,repository ./...
```

