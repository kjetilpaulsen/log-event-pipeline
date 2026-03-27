# log-event-pipeline

*This is a personal learning project for JS and GO. I find that playing around in a codebase is better for learning than starting from scratch. As a result, parts of the initial code was written by an LLM. There will be significant changes to this project the coming days and weeks. I will update the GitHub issues page with planned features as I go.*

### Description

A real-time log event pipeline that collects structured JSON log events over TCP, filters high-severity events, and streams them to a browser dashboard via Server-Sent Events (SSE).

### What this is

This repository is a **starting point**, not a finished application.

It provides:
- A TCP server for ingesting structured log events from any application
- An SSE-based HTTP server for streaming high-severity events to connected clients
- A browser dashboard for viewing live ERROR and CRITICAL events
- A Python log generator for local testing and development

## Table of contents
- [Quick start](#quick-start)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
  - [Requirements](#requirements)
  - [Setup](#setup)
  - [Usage](#usage)
- [Log generator (Python)](#log-generator-python)

## Quick Start

```bash
git clone ...
cd log-event-pipeline
go run ./cmd/server
```

Open your browser at `http://127.0.0.1:9001` to see the dashboard.

## Features

- TCP server for receiving newline-delimited JSON log events
- Severity filter: only ERROR and CRITICAL events are forwarded to the dashboard
- Server-Sent Events (SSE) endpoint for real-time browser streaming
- Lightweight frontend (vanilla JS) showing live event cards with level, timestamp, app name, function, and message
- Maximum 50 events displayed, with oldest auto-removed
- Health check endpoint (`/health`)
- Python log generator for integration testing and local development
- Concurrent client support with safe channel-based broadcasting

## Architecture

The pipeline has three components:

```
[App / Log Generator] --TCP (port 9000)--> [Go Server] --SSE (port 9001)--> [Browser Dashboard]
```

- **TCP ingest** (`127.0.0.1:9000`): Receives newline-delimited JSON `LogEvent` objects from any connected client
- **Broker**: Filters events by severity and broadcasts ERROR/CRITICAL events to all connected SSE clients
- **HTTP server** (`127.0.0.1:9001`): Serves the frontend and the `/events` SSE stream

### Log event schema

```json
{
  "timestamp": "2025-01-01T00:00:00Z",
  "level": "ERROR",
  "app_name": "myapp",
  "logger_name": "myapp.module",
  "function_name": "do_something",
  "message": "Something went wrong"
}
```

## Installation

### Requirements

- **Go** ≥ 1.21
- **Node.js** + **npm** (for the frontend, if editing or bundling JS)
- **Python** ≥ 3.11 (optional, for the log generator)

### Setup

```bash
git clone https://github.com/<your-username>/log-event-pipeline.git
cd log-event-pipeline
```

For the Go server, no additional dependencies are needed — the server uses only the standard library:

```bash
go build ./...
```

For the frontend (if you want to install dev tooling or a bundler):

```bash
npm install
```

### Usage

**Start the server:**

```bash
go run ./cmd/server
```

Or build and run the binary:

```bash
go build -o server ./cmd/server
./server
```

The server listens on two addresses by default:
- `127.0.0.1:9000` — TCP ingest (log events in)
- `127.0.0.1:9001` — HTTP dashboard and SSE stream (events out)

**Check the health endpoint:**

```bash
curl http://127.0.0.1:9001/health
```

**View the dashboard:**

Open `http://127.0.0.1:9001` in your browser. Any ERROR or CRITICAL events received over TCP will appear in real time.

**Connect via SSE directly:**

```bash
curl -N http://127.0.0.1:9001/events
```

## Log Generator (Python)

A Python script (`main.py`) is included to generate and send random log events to the TCP server. It requires no additional packages beyond the standard library.

**Run with defaults (10 000 events, sent as fast as possible):**

```bash
python main.py
```

**Run in dev mode (slower, with a delay between events):**

```bash
python main.py --dev
```

**Available options:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dev` | false | Add a 0.25s delay between events |
| `--logs` | 10000 | Number of events to generate |
| `--host` | 127.0.0.1 | Target host |
| `--port` | 9000 | Target port |

**Example with custom options:**

```bash
python main.py --dev --logs 100 --host 127.0.0.1 --port 9000
```

Only ERROR and CRITICAL events (roughly 2/5 of generated events) will appear in the browser dashboard, since the server filters lower-severity levels.

