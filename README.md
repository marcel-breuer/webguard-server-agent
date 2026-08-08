# WebGuard Server Agent

WebGuard Server Agent is a privacy-focused, outbound-only Linux agent for reporting bounded server telemetry to a WebGuard Server Health monitoring.

It is intended for customer-managed servers and reports CPU, memory, load, uptime, and optional local service-check results. The agent does not expose an inbound listener, execute remote commands, or collect filesystem, process, environment, or customer-payload data.

The project is in its initial implementation phase. The repository now provides reproducible Linux build and release foundations; the reporting implementation and its API contract are tracked in [WebGuard Server Agent #2](https://github.com/marcel-breuer/webguard-server-agent/issues/2) and [WebGuard Core #658](https://github.com/marcel-breuer/webguard/issues/658).

## Development

Use Go 1.24 or the included Dockerfile. Run the complete quality suite with:

```sh
make check
make dist VERSION=dev
```

Version tags publish Linux `amd64` and `arm64` binaries together with SHA-256 checksums. See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor and security guidance.

## Contributing

Contributions are welcome. Please start with the repository guidance and open an issue before beginning larger changes.

## License

This project is licensed under the [MIT License](LICENSE).
