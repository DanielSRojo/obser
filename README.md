# Obser - Obsidian CLI Tool

Obser is a command-line interface (CLI) tool for managing and analyzing Obsidian notes. It provides various commands to list and analyze notes, journal entries, and properties from your Obsidian vault.

## Features

- List all notes in your vault
- List journal entries
- List all properties used in notes
- Generate statistics about your notes
- Analyze monthly and yearly activity

## Installation

1. Install Go (version 1.18 or higher)
2. Clone this repository
3. Build the project:

   ```bash
   go build -o obser
   ```

4. Move the binary to your PATH:

   ```bash
   sudo mv obser /usr/local/bin/
   ```
   For user installation:
   ```bash
   sudo mv obser /usr/local/bin/
   ```


## Usage

```bash
obser [command]
```

The vault is read from `--vault`, then the `OBSER_VAULT` environment variable, and defaults to `~/.obsidian`:

```bash
obser --vault ~/notes journal entries
```

### Available Commands

#### List Commands

```bash
obser list [subcommand]
```

- `notes`: List all notes in the vault

  ```bash
  obser list notes
  ```

- `journal`: List all journal entries

  ```bash
  obser list journal
  ```

- `properties`: List all properties used in notes

  ```bash
  obser list properties
  ```

#### Statistics Commands

:> [!WARNING]

```bash
obser journal statistics [year] [month]
```

- Get monthly statistics:

  ```bash
  obser journal statistics 2023 10
  ```

- Get yearly statistics:

  ```bash
  obser journal statistics 2023
  ```

#### Fill Command

```bash
obser journal fill [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--dry-run]
```

Creates the missing journal entries from the template configured in Obsidian's Daily notes plugin (`.obsidian/daily-notes.json`), rendering its Templater date expressions for each day. By default it fills every day after the last entry up to today. Existing entries are never overwritten.

- Catch up since the last entry:

  ```bash
  obser journal fill
  ```

- Preview filling a specific range:

  ```bash
  obser journal fill --from 2025-12-28 --to 2026-01-03 --dry-run
  ```

## Configuration

You can specify a configuration file using the `--config` flag:

```bash
obser --config /path/to/config.toml [command]
```
