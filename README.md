# providers-go-sdk

Go API clients for cloud providers that don't publish an official Go SDK, used by [Raklane](https://raklane.com) to provision servers. Each provider is its own package; they share a small HTTP layer in `internal/httpx`.

| Package | Provider | Covers |
|---|---|---|
| `idcloudhost` | [IDCloudHost](https://idcloudhost.com) (Indonesia) | VMs, block storage, private networks, floating IPs, firewalls, locations, images, billing accounts, pricing |

The clients follow these rules:

- **Plain API clients.** They have no business logic; mapping provider resources onto your own model is up to the caller.
- **Context on every call.** `HTTPClient` and `BaseURL` can be overridden, so tests run against `httptest`.
- **Typed errors.** A non-2xx answer is an `*APIError` with `NotFound()` and `Throttled()`, so callers can treat "already gone" as success.
- **Only GETs are retried** on 429/5xx. A create whose answer was lost may still have happened, so the caller decides how to recover from it.
- **Standard library only.**

## IDCloudHost

```go
c := idcloudhost.New(apiKey, "jkt01") // "" = the account's default location
c.BillingAccountID = 123

vm, err := c.CreateVM(ctx, idcloudhost.CreateVMRequest{
	Name: "web-1", 
	OSName: "ubuntu", 
	OSVersion: "24.04",
	DiskGB: 40, 
	VCPU: 2, 
	RAMMB: 4096,
	Username: "admin", 
	Password: "Generated-Secret-1",
	CloudInit: "#cloud-config\nruncmd:\n  - echo hello\n",
})
if idcloudhost.IsNotFound(err) { /* ... */ }
```

The endpoints follow IDCloudHost's API reference at https://api.idcloudhost.com.

## Layout

The repository follows the standard Go module layout (https://go.dev/doc/modules/layout): one importable package per provider at the module root, and shared code under `internal/`, which other modules can't import.

```
providers-go-sdk/
├── idcloudhost/          package idcloudhost
│   ├── doc.go            package documentation
│   ├── client.go         Client, New, request plumbing
│   ├── errors.go         APIError, IsNotFound
│   ├── vm.go             one file per API resource:
│   ├── disk.go             vm, disk, network, floatingip, firewall,
│   ├── ...                 location, image, billing, pricing
│   ├── *_test.go         tests next to the file they cover
│   └── example_test.go   runnable examples shown on pkg.go.dev
└── internal/
    └── httpx/            shared HTTP layer: bodies, errors, retries
```

To add a provider, create a new top-level package (e.g. `./hetzner`) with the same shape, and build it on `internal/httpx`.

## Development

```sh
go test -race ./...
```

To work on this module together with a project that uses it, use an uncommitted `go.work` in the parent directory:

```
go 1.25.0

use (
	./providers-go-sdk
	./your-project
)
```

### Releases

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) checks formatting, vets and runs the tests on every push and pull request. When they pass on `main`, it tags the commit with the next version, picked from the commit messages since the last tag ([Conventional Commits](https://www.conventionalcommits.org)):

| Commits since the last tag | Bump | Example |
|---|---|---|
| `feat:` or a breaking change (`feat!:`, `BREAKING CHANGE`) | minor | v0.3.2 → v0.4.0 |
| anything else (`fix:`, `docs:`, ...) | patch | v0.3.2 → v0.3.3 |

The first tag is `v0.1.0`. Major versions are tagged by hand, because Go needs a `/vN` suffix on the module path for v2 and later. After a release, bump the version in the consuming project's `go.mod`.
