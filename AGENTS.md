# Repository Guidance

Read [CONTRIBUTING.md](CONTRIBUTING.md) before changing this repository.

- Prefer the Dockerfile for builds and run `make check` before submitting changes.
- Keep the Linux targets to `amd64` and `arm64` unless release support is explicitly expanded.
- Preserve the outbound-only and least-privilege design; never add remote command execution or telemetry beyond the documented scope.
- Keep credentials, report URLs, local configuration, and customer data out of version control and logs.
- Keep pull requests focused, with tests and documentation updated when behaviour changes.

## CI/CD-Pipeline

Stand: 2026-11-15

Behalten:
- `.github/workflows/quality.yml`: Unit-Tests (`make test`) bei Pull Requests und Pushes auf `main`.
- `.github/workflows/release.yml`: Tag-Erstellung (`create-tag`, nächster Patch-Tag) und Release-Veröffentlichung (`release`, inkl. Quellprüfung `make check`, Builds und `SHA256SUMS`) bei Pushes auf `main`.

Entfernt:
- Dependabot-Konfiguration `.github/dependabot.yml`: Dependency-Updates, nicht nötig für die Pipeline.
- `.github/workflows/dependabot-auto-merge.yml`: Auto-Merge für Dependabot-PRs, nicht nötig für die Pipeline.
- Im Quality-Workflow entfernt: Format-/Vet-Prüfung (`make check` ohne Tests), Build der Linux-Artefakte und Debian-Pakete. Diese Schritte prüfen Stil bzw. Build, nicht die Funktion.

Hinweis: Browser-/E2E-Tests existieren im Repo nicht; es gibt keinen Job dafür.

Regel: Vor jedem PR lokal alle Pipeline-Schritte ausführen (`make test`, sowie für Releases `make check`, `make dist VERSION=dev`, `make deb VERSION=v0.0.0`).
