# WebGuard Server Agent

WebGuard Server Agent is a privacy-focused, outbound-only Linux agent for reporting bounded server telemetry to a WebGuard Server Health monitoring.

It is intended for customer-managed servers and reports CPU, memory, load, uptime, and optional local service-check results. The agent does not expose an inbound listener, execute remote commands, or collect filesystem, process, environment, or customer-payload data.

The agent sends the versioned Server Health report defined by the [Core API contract](https://github.com/marcel-breuer/webguard/blob/main/docs/integrations/server-agent-api.md). It has no inbound listener, accepts no remote commands, and does not collect filesystem, process, environment, mount, cloud-metadata, or customer-payload data.

## Install and configure

Download a verified release for Linux `amd64` or `arm64`, verify its SHA-256 checksum, then unpack it on the customer server. Run the included installer as root:

```sh
sudo ./scripts/install.sh ./webguard-server-agent
sudoedit /etc/webguard-server-agent/config.json
sudo systemctl enable --now webguard-server-agent
```

The report URL is the private Server Health URL displayed in the WebGuard monitoring detail page. It is the only credential needed by the agent and must never be shared. The installer creates the dedicated `webguard-server-agent` system user, a root-owned configuration directory, and a state directory owned by that user.

Use [config.example.json](config.example.json) as the configuration reference. The default report interval is 60 seconds. Optional service checks must target `localhost` or a loopback IP address; their local targets, headers, and credentials never leave the server.

```sh
sudo webguard-server-agent --status
sudo webguard-server-agent --once
sudo journalctl -u webguard-server-agent -f
```

For an upgrade, stop the unit, replace `/usr/local/bin/webguard-server-agent` with a checksum-verified release binary, and start it again. To roll back, repeat those steps with the prior verified binary. The on-disk queue is retained across restarts. To uninstall, disable the unit, remove the binary and unit file, then deliberately remove `/etc/webguard-server-agent` and `/var/lib/webguard-server-agent` only after retaining any configuration or queue data you need.

## Reliability and privacy

The agent validates TLS, retries connection failures, timeouts, HTTP `429`, and `5xx` responses with bounded exponential backoff and jitter, then retains reports in a size- and age-limited local queue. Authentication and validation failures are not retried. Logs redact the report URL and never include its token.

The installed systemd unit uses a dedicated unprivileged user and hardening settings. It only permits outbound networking and access to its designated state directory.

## Development

Use Go 1.24 or the included Dockerfile. Run the complete quality suite with:

```sh
make check
make dist VERSION=dev
```

Every merge to `main` creates the next patch version tag and publishes Linux `amd64` and `arm64` binaries together with SHA-256 checksums. See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor and security guidance.

## Contributing

Contributions are welcome. Please start with the repository guidance and open an issue before beginning larger changes.

## License

This project is licensed under the [MIT License](LICENSE).
