# Contributing

Thank you for contributing to WebGuard Server Agent. Please discuss substantial changes in an issue before opening a pull request.

## Development

The project uses Go and supports Linux `amd64` and `arm64` release artifacts. Use the repository Dockerfile for a production-like build:

```sh
docker build -t webguard-server-agent .
```

Run the complete quality suite before opening a pull request:

```sh
make check
make dist VERSION=dev
```

Keep changes focused and add tests for behaviour changes. Run `gofmt` on changed Go files.

## Security and privacy

Do not commit report URLs, tokens, customer addresses, credentials, local configuration, or captured telemetry. The package must remain outbound-only, verify TLS by default, avoid remote execution, and collect only the documented metrics.

Report security issues privately to the maintainers rather than publishing sensitive details in an issue.

## Releases

Every merge to `main` creates the next patch version tag and triggers a release build for both supported Linux architectures. The release publishes standalone binaries, Debian packages, and a signed APT repository. Release artifacts include a `SHA256SUMS` file. Validate the checksums before publishing installation instructions.

The repository signing key is managed only through encrypted GitHub Actions secrets. Key rotation requires replacing the public key in `packaging/apt/` and publishing updated repository metadata before changing installation instructions.
