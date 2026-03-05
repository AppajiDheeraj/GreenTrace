# CarbonQt

[![CI](https://img.shields.io/github/actions/workflow/status/AppajiDheeraj/GreenTrace/ci.yml?branch=main&logo=github)](https://github.com/AppajiDheeraj/GreenTrace/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AppajiDheeraj/GreenTrace?logo=go)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/AppajiDheeraj/GreenTrace?logo=github)](https://github.com/AppajiDheeraj/GreenTrace/releases)

CarbonQt is a cross-platform CLI dashboard for estimating process energy use and carbon emissions in real time. It provides a clean TUI, highlights top emitters, and supports quick actions like process selection and termination.

## Features

- Live system overview (CPU, RAM, platform, uptime)
- Top carbon process summary
- Process table with CPU, memory, power, carbon, runtime, and path
- Keyboard navigation with a kill action
- Repo-aware process filtering

## Quick Start

```bash
go build -o carbonqt
./carbonqt dashboard
```

## Commands

- `carbonqt dashboard` - launch the interactive dashboard
- `carbonqt run 10s` - monitor for a fixed duration and print a report

## Flags

- `--repo-only` (default: true) - restrict process list to the current repository
- `--cpu-tdp` - CPU TDP in watts (default: 65)
- `--emission-factor` - kg CO2 per joule (default: 2e-10)

## Controls (Dashboard)

- Up/Down - select process
- `K` - kill selected process
- `Q` - quit

## Notes

- On some systems, killing processes may require elevated permissions.
- Process paths can be long and are truncated in the table for readability.

## Development

```bash
make fmt
make test
make build
```

## Release and Tagging

Build release artifacts locally:

```bash
./scripts/build-release.sh
```

On Windows PowerShell:

```powershell
./scripts/build-release.ps1
```

For release notes, use GitHub Releases and summarize:
