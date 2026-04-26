# term

`term` is an offline developer terminal assistant CLI. It helps answer:

- How do I run this project?
- What command should I use?
- Why did this command fail?
- What command did I run before?

No AI integration is included in this MVP, and no external APIs are called.

## Install From Source

```sh
go install .
```

Or build a local binary:

```sh
go build -o term .
```

## Quick Start

```sh
term init
term list
term run dev
```

`term init` creates a `.term.yml` in the current directory. It detects common project types such as Node.js, Maven Spring Boot, Docker Compose, frontend folders, .NET, and Go.

## Project Shortcuts

Example `.term.yml`:

```yaml
project: payment-service

shortcuts:
  dev:
    description: Start local development
    steps:
      - npm run dev

  test:
    description: Run tests
    steps:
      - npm test

  add-migration:
    description: Add EF Core migration
    args:
      - name
    steps:
      - dotnet ef migrations add {{name}}
```

Run shortcuts:

```sh
term run dev
term run test
term run add-migration CreateUsersTable
```

Steps can be strings:

```yaml
steps:
  - npm run dev
```

Or objects:

```yaml
steps:
  - name: frontend
    command: cd frontend && npm run dev
```

Parallel shortcuts are supported:

```yaml
shortcuts:
  dev:
    description: Start local development
    parallel: true
    steps:
      - name: backend
        command: ./mvnw spring-boot:run
      - name: frontend
        command: cd frontend && npm run dev
```

Dangerous shortcuts can require confirmation:

```yaml
shortcuts:
  reset-db:
    description: Reset local database
    danger: high
    confirm: true
    steps:
      - docker compose down -v
      - docker compose up -d db
```

## Command Search

Search built-in command recipes:

```sh
term find "kill port 8080"
term find "undo last commit"
```

Recipes include Git, Docker, Java/Spring, Node/React, PostgreSQL, Linux/macOS, and Windows commands. Port numbers in queries replace `<PORT>` placeholders automatically.

## Error Explanation

Explain common errors:

```sh
term explain "EADDRINUSE: address already in use :::3000"
echo "EADDRINUSE: address already in use :::3000" | term explain
```

Use the latest failed recorded command:

```sh
term explain last
```

## History Recording

Record commands manually:

```sh
term record --command "npm run dev" --exit-code 1 --stderr "EADDRINUSE: address already in use :::3000"
```

History is stored in SQLite at:

```text
~/.term/term.db
```

Search, show, rerun, or clear history:

```sh
term history search postgres
term history show 1
term history run 1
term history clear
```

## Shell Hook Placeholder

Print the experimental hook snippet:

```sh
term hook install
```

The MVP does not install deep shell integration yet. Future versions will use hooks to record commands automatically.

## Roadmap

- Safer automatic shell history capture
- Better command templates and variable prompts
- Project-specific recipe packs
- Richer diagnostics for build tools
- Optional local-only AI integration after the offline MVP is stable
