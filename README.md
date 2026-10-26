# Employee CRUD with Go and Gin

Iteration 1: employees live in a concurrency-safe **slice** (Go's list-like collection). The server starts empty. Data disappears when the process stops.

## Run

From PowerShell:

```powershell
cd Z:\Study\employee-api
go run ./cmd/api
```

Open http://127.0.0.1:8080/healthz. The module requires Go 1.25+; Go 1.21+ can automatically download the requested toolchain when GOTOOLCHAIN=auto. Gin is pinned to v1.12.0; commit go.mod and go.sum together. The module name example.com/employee-api is a local placeholder; replace it and imports when publishing to your own repository.

Optional configuration:

```powershell
$env:HTTP_ADDR = '127.0.0.1:9090'
$env:GIN_MODE = 'release'
go run ./cmd/api
```

Ctrl+C shuts the server down gracefully. No .env loader is required; configuration comes from environment variables. This learning API has no authentication and binds to localhost by default.

## Try the CRUD flow

```powershell
$base = 'http://127.0.0.1:8080/api/v1/employees'
$body = @{ name = 'Anand'; email = 'anand@example.com'; department = 'Engineering' } | ConvertTo-Json
$created = Invoke-RestMethod -Method Post -Uri $base -ContentType 'application/json' -Body $body
$id = $created.data.id
Invoke-RestMethod -Uri $base
Invoke-RestMethod -Uri "$base/$id"
$update = @{ name = 'Anand T'; email = 'anand@example.com'; department = 'Platform' } | ConvertTo-Json
Invoke-RestMethod -Method Put -Uri "$base/$id" -ContentType 'application/json' -Body $update
Invoke-RestMethod -Method Delete -Uri "$base/$id"
```

| Method | Route | Success |
|---|---|---|
| POST | /api/v1/employees | 201, Location header, created employee |
| GET | /api/v1/employees | 200, array (empty is []) |
| GET | /api/v1/employees/:id | 200, employee |
| PUT | /api/v1/employees/:id | 200, replaced employee |
| DELETE | /api/v1/employees/:id | 204, no body |

PUT requires all three fields. IDs are generated on the server, monotonically increase, and are never reused during a process lifetime. Emails are normalized and unique regardless of case. Name and department are trimmed and required (up to 100 bytes); email is validated (up to 254 bytes). Unknown JSON fields and multiple JSON values are rejected; bodies are limited to 1 MiB. Invalid input gets 400, oversized bodies 413, wrong media type 415, absent employees 404, duplicate emails 409. Unexpected errors are logged while clients receive a generic 500.

Success envelope: `{"data": ...}`. Error envelope: `{"error":{"code":"not_found","message":"employee not found"}}`.

## Project structure and dependency injection

```text
cmd/api/main.go                   Composition root, configuration, server lifecycle
internal/employee/employee.go      Domain model
internal/employee/employee_request.go Create/update request
internal/employee/service.go       Business operations
internal/employee/validation.go    Normalization and validation
internal/employee/repository.go    Storage contract
internal/employee/errors.go        Domain errors
internal/repository/employee_list.go      In-memory slice implementation
internal/httpapi/router.go         Middleware and resource route registration
internal/httpapi/employee_handler.go Employee CRUD endpoints
internal/httpapi/employee_routes.go Employee route registration
internal/httpapi/employee_errors.go Employee-specific error mapping
internal/httpapi/request.go        Shared JSON decoding and ID parsing
internal/httpapi/response.go       Shared error responses
internal/httpapi/health_handler.go Health endpoint
internal/httpapi/employee_response.go Employee response DTO and mapping
docs/architecture.md               Design walkthrough and database migration plan
```

Request flow: HTTP handler -> employee service -> repository -> slice.

**Yes, Go supports both dependency injection and the repository pattern.** Dependency injection usually means passing dependencies into constructors. You do not need a framework or annotations:

```go
repo := repository.NewEmployeeList()
service := employee.NewService(repo)
router := httpapi.NewRouter(service, log)
```

The service accepts a Repository interface, so it does not know where data is stored. Go interfaces are satisfied implicitly; the list repository needs no implements keyword. The HTTP handler accepts the concrete service because there is no current need for a second service implementation. Avoid adding an interface for every struct.

## Verify

```powershell
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/api
# Optional: requires a supported C toolchain for Go's race detector on Windows.
go test -race ./...
```

Tests exercise the real router, service, and list repository together, including CRUD, normalization, duplicate emails, invalid input, and missing records. Repository tests cover concurrent access, copy isolation, uniqueness, and cancellation.

## Learning sequence

1. Run the PowerShell CRUD example and inspect requests/responses.
2. Follow Create from handler to service to repository.
3. Change validation in the service and add a test.
4. Implement a database repository using docs/architecture.md.

References: [Effective Go](https://go.dev/doc/effective_go), [Go module layout](https://go.dev/doc/modules/layout), [Gin](https://github.com/gin-gonic/gin).

Request and response naming: use EmployeeRequest for create/update payloads and EmployeeResponse for returned employee data. These are Go structs (Go does not have classes). Employee remains the internal domain model.




## List endpoint rate limit

GET /api/v1/employees allows at most 10 requests per client IP in any rolling 5-second window. Additional requests return HTTP 429 with the usual error envelope (code rate_limit_exceeded) and Retry-After in whole seconds, rounded up. Blocked requests do not extend the window. Other endpoints, including GET /api/v1/employees/:id, are unaffected.

The reusable middleware lives in internal/httpapi/rate_limit.go and is attached in employee_routes.go. It synchronizes concurrent requests, expires inactive client entries lazily, and stores quotas in memory per server process. Restarting resets quotas; multiple replicas would require a shared limiter. Trusted proxies are disabled, so clients cannot change quotas by supplying X-Forwarded-For. If deploying behind a reverse proxy, configure its trusted addresses deliberately so client IPs resolve correctly. Clients sharing one IP also share one quota.


## Swagger API documentation

Start the server with go run ./cmd/api, then open http://127.0.0.1:8080/swagger/index.html. Swagger UI lets you inspect schemas and try all CRUD endpoints. The JSON specification is at /swagger/doc.json. Both URLs use the same server address as the API; no host or port is hardcoded in the specification.

Annotations next to handlers are the source of truth. After changing API annotations or request/response schemas, run:

```powershell
go generate ./internal/httpapi
```

The generator is pinned to Swag v1.16.6. It writes internal/apidocs/docs.go, swagger.json, and swagger.yaml; commit these files together with annotation changes. Normal startup needs no generation step or external CDN. The generated specification uses Swagger 2.0. The UI is served separately from employee handlers and is not subject to the employee-list rate limit; requests made through Try it out still follow normal endpoint limits.
