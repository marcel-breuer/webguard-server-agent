# WebGuard Server Agent

WebGuard Server Agent is a privacy-focused, outbound-only Linux agent for reporting bounded server telemetry to a WebGuard Server Health monitoring.

It is intended for customer-managed servers and reports CPU, memory, load, uptime, and optional local service-check results. The agent does not expose an inbound listener, execute remote commands, or collect filesystem, process, environment, or customer-payload data.

The agent sends the versioned Server Health report defined by the [Core API contract](https://github.com/marcel-breuer/webguard/blob/main/docs/integrations/server-agent-api.md). It has no inbound listener, accepts no remote commands, and does not collect filesystem, process, environment, mount, cloud-metadata, or customer-payload data.

## Install and configure

### Debian and Ubuntu package

Download the matching `.deb` package from the [GitHub release](https://github.com/marcel-breuer/webguard-server-agent/releases), verify it against `SHA256SUMS`, and install it:

```sh
sudo apt install ./webguard-server-agent_VERSION_ARCH.deb
sudoedit /etc/webguard-server-agent/config.json
sudo systemctl restart webguard-server-agent
```

The package installs the systemd unit, creates the dedicated `webguard-server-agent` user, enables the service, and preserves configuration and queued reports during upgrades. Remove the package with `sudo apt-get remove webguard-server-agent`; use `sudo apt-get purge webguard-server-agent` only when you also intend to remove the configuration and queued reports.

### Manual installation

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

For a manual upgrade, stop the unit, replace `/usr/bin/webguard-server-agent` with a checksum-verified release binary, and start it again. To roll back, repeat those steps with the prior verified binary. The on-disk queue is retained across restarts. To uninstall, disable the unit, remove the binary and unit file, then deliberately remove `/etc/webguard-server-agent` and `/var/lib/webguard-server-agent` only after retaining any configuration or queue data you need.

## Reliability and privacy

The agent verifies the report server using system certificate roots by default, supports an additional CA file, requires TLS 1.2 or newer when a custom CA is configured, and does not follow redirects. Reports are limited to 64 KiB, response bodies to 64 KiB, and each HTTP request to 15 seconds. Connection failures, timeouts, HTTP `429`, and `5xx` responses are retried at most 10 times with exponential backoff and jitter; the configured maximum delay is five minutes. Authentication, validation, redirect, and oversized-response failures are not retried. Logs redact the report URL and never include its token.

The on-disk queue retains reports for the configured maximum age, up to 64 MiB and 4,096 reports. When it reaches either limit, it removes the oldest reports to make room. Each stored report is capped at 64 KiB. `--status` reads only local queue state and shows its report count, byte usage, and oldest report age; it does not contact the report endpoint. Host resource metrics are the documented report fields sent to Server Health, with no additional host inventory collected.

## Collection and report contract

Linux collection reads only `/proc/stat`, `/proc/loadavg`, `/proc/meminfo`, and `/proc/uptime`; the collector has no transport dependency. Optional service checks are limited to 20 configured loopback HTTP or TCP targets, each with a timeout of at most 30 seconds. HTTP checks do not follow redirects or read response bodies. The default sampling interval is 60 seconds and the minimum is 10 seconds.

Reports use `schema_version: 1` and the field names represented by [the Core v1 fixture](testdata/core-v1-report.json). Compatible additions are optional JSON fields, which v1 decoders ignore. A change that renames or changes the meaning of an existing field requires a new schema version and a coordinated Core API rollout. Queued reports retain their original sample time and report ID across retries.

To measure collection and delivery costs under constrained resources, run the benchmarks in the project container with one CPU and a 128 MiB memory limit:

```sh
make bench
```

The collection benchmark uses representative `/proc` samples and the delivery benchmark posts to a local test endpoint. Run them on the intended low-resource host class before changing sampling, retry, or queue defaults.

The installed systemd unit uses a dedicated unprivileged user and hardening settings. It only permits outbound networking and access to its designated state directory.

## Development

Use Go 1.24 or the included Dockerfile. Run the complete quality suite with:

```sh
make check
make dist VERSION=dev
```

Every merge to `main` creates the next patch version tag and publishes Linux `amd64` and `arm64` binaries, Debian packages, and SHA-256 checksums. See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor and security guidance.

## Contributing

Contributions are welcome. Please start with the repository guidance and open an issue before beginning larger changes.

## License

This project is licensed under the [MIT License](LICENSE).
