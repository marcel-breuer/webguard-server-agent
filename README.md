# WebGuard Server Agent

WebGuard Server Agent is a privacy-focused, outbound-only Linux agent for reporting bounded server telemetry to a WebGuard Server Health monitoring.

It is intended for customer-managed servers and reports CPU, memory, load, uptime, and optional local service-check results. The agent does not expose an inbound listener, execute remote commands, or collect filesystem, process, environment, or customer-payload data.

The project is in its initial setup phase. Its API contract and WebGuard Core integration are tracked in [WebGuard Core #658](https://github.com/marcel-breuer/webguard/issues/658).

## Contributing

Contributions are welcome. Please start with the repository guidance and open an issue before beginning larger changes.

## License

This project is licensed under the [MIT License](LICENSE).
