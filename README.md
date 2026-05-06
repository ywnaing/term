# term

`term` is an offline developer terminal assistant CLI. It helps answer:

- How do I run this project?
- What command should I use?
- Why did this command fail?
- What command did I run before?

No AI integration is included in this MVP, and no external APIs are called.

## Install From Source

Install a tagged release with Go:

```sh
go install github.com/ywnaing/term@v0.1.0
```

Or install from a local checkout:

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
term doctor
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

## Shell Hook

Install the experimental shell hook:

```sh
term hook install
term hook install zsh
term hook install bash
```

`term hook install` edits `~/.zshrc` or `~/.bashrc` with a managed block and creates a timestamped backup first. Re-running the command updates the existing managed block instead of adding duplicates.

Preview or print the hook without changing files:

```sh
term hook install zsh --dry-run
term hook install zsh --print
```

Remove the managed hook:

```sh
term hook uninstall zsh
term hook uninstall bash
```

Check hook installation:

```sh
term hook status
term hook status zsh
```

The hook records command text, exit code, cwd, project name, timestamp, shell, OS, and duration after each command. It does not capture stdout or stderr by default, which keeps history safer for everyday use.

Commands that start with a space are skipped, and `term record` redacts common secret values such as tokens, passwords, API keys, and bearer tokens before saving history.

To disable automatic recording in a project, add this to `.term.yml`:

```yaml
history:
  enabled: false
```

`term explain last` works best when stderr was recorded manually with `term record --stderr`, while automatic hooks make `term history search` useful immediately.

## Roadmap

- Better `term doctor` suggestions for project-specific tools
- Optional stderr capture with strong redaction controls
- Better command templates and variable prompts
- Project-specific recipe packs
- Richer diagnostics for build tools
- Optional local-only AI integration after the offline MVP is stable
