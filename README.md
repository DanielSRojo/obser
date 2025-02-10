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

## Usage

```bash
obser [command]
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
obser statistics [year] [month]
```

- Get monthly statistics:

  ```bash
  obser statistics 2023 10
  ```

- Get yearly statistics:

  ```bash
  obser statistics 2023
  ```

## Configuration

You can specify a configuration file using the `--config` flag:

```bash
obser --config /path/to/config.toml [command]
```
