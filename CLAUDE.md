# Repository Guidance

Read [CONTRIBUTING.md](CONTRIBUTING.md) before changing this repository.

- Prefer the Dockerfile for builds and run `make check` before submitting changes.
- Keep the Linux targets to `amd64` and `arm64` unless release support is explicitly expanded.
- Preserve the outbound-only and least-privilege design; never add remote command execution or telemetry beyond the documented scope.
- Keep credentials, report URLs, local configuration, and customer data out of version control and logs.
- Keep pull requests focused, with tests and documentation updated when behaviour changes.
