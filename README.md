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

#### Tidy Command

```bash
obser journal tidy [--before YYYY-MM-DD] [--dry-run]
```

Removes the frontmatter properties that were never filled in (`key:`, `null`, `~`, `""`, `[]` or `{}`) from the journal entries dated before today, or before the day given with `--before`. Only `YYYY-MM-DD.md` notes with `type: journal` are touched. Today's and future entries keep their template prompts, and `type`, `false`, and `0` values are always kept.

This command edits your notes, so keep the vault in git or a sync service, and preview first:

```bash
obser journal tidy --dry-run
obser journal tidy
```

Each changed note is listed with the properties removed from it, followed by a summary. Only the lines of the empty properties are deleted; the rest of the note stays byte-for-byte the same. Files are rewritten in place and keep their modification time, so Dataview `file.cday`/`file.mday` queries are unaffected. Entries whose frontmatter can't be parsed, or that still contain Templater tags, are skipped with a warning, and the command exits non-zero.

- Only tidy entries from before 2025:

  ```bash
  obser journal tidy --before 2025-01-01 --dry-run
  ```

### Habits

Every frontmatter property of a journal entry that is done at least once (a number, a duration like `30 minutes`, or `true`) is a habit. A habit is done on a day when its value is `true`, a number above zero, or any non-empty text; an empty value, `false`, or a missing entry is not done. Rates only count days from the first day a habit was done, up to today.

Journal entries are the `YYYY-MM-DD.md` notes (with or without `type: journal`) directly inside the folder configured in Obsidian's Daily notes plugin, or the vault root when none is set. Durations can combine hours, minutes and seconds, with decimals and short units (`1 hour 30 minutes`, `1.5 hours`, `90 min`, `1h30m`; single-letter units must touch the number, so `500 m` stays meters) and are added up in seconds, and a bare number takes the habit's unit when it only uses one (`journaling: 120` next to `150 words` counts as 120 words); on a habit measured in time, a bare number is minutes (`guitar: 25` counts as 25 minutes). When a habit mixes units, only the most frequent one is added up and a warning is printed.

#### Summary

```bash
obser habits [--year YYYY] [--month M]
```

Shows one row per habit for the last 365 days, for a calendar year with `--year`, or for a calendar month with `--month` (of `--year`, or of the current year): days done, completion rate, current streak, longest streak, total, average per done day, and a sparkline of the completion rate over the last 12 months.

```bash
obser habits
obser habits --year 2024
obser habits --year 2024 --month 3
```

```
HABIT         DAYS  RATE  STREAK  BEST      TOTAL        AVG  TREND
calisthenics    58   16%       0     1    37h 50m     39m 8s  ▂▂▂▂▁▂▃▃▂▂▂▂
chain          271   74%       5    18          -          -  ▆▅▅▆▇▇▆▆▆▆▆▇
guitar         118   32%       0     9    51h 56m    26m 24s  ▅▂▄▄▂▂▅▃▂▁▃▄
journaling       4    1%       0     1  410 words  103 words  ▁▁▁▁▁▁▁▁▁▁▁▁
```

Colors are used only when the output is a terminal and `NO_COLOR` is not set.

> [!NOTE]
> `obser journal statistics` was removed. Its totals are the `TOTAL` column (or `DAYS`, for habits without amounts like `chain: true`) of `obser habits --year YYYY [--month M]`.

#### Single habit

```bash
obser habits show <habit> [--year YYYY]
```

Shows one habit for a year (the current year by default): a header with its days done, rate, streaks, total and average, a calendar heatmap of every day, one bar per month with its days done (and total, for habits with amounts), and the completion rate on each weekday. In the heatmap, done days are shaded by quartile of the habit's amounts that year (booleans are always the darkest), and days before the habit started or after today are left blank.

```bash
obser habits show guitar --year 2025
```

```
guitar · 2025
Days 167 · Rate 46% · Streak 0 · Best 20 · Total 90h 30m · Avg 32m 31s

    Jan Feb Mar  Apr May Jun  Jul Aug  Sep Oct Nov  Dec
Mon  ████·▒░█▓▒·█▓▓··▓·▒██░·······▒······░·███▒▒·▒░▒·▒···
     ▓▒▒▒░▓░░▒▒▒░░▒░▒█··██░·░····▒░········▓█░█▒▓·▒······
Wed ░▓▒█▒·█·▒··▓·░·░░░··▓▓·······▓▒······░░░▓▒░▓░░░░▓·▒··
    ░▓██·█░▓▒··░·▓·▓··█·░········█········█·▒░▒▒·▒···▒··
Fri ·▒▒·▒·▓░·░·░▓··▓··▓··▓········▒···░······░·▓··░·····
    ▒░░▒··░··▒·░▒·····▓·▓·░······░··········▓░▓▒·▒·▒·▒··
    █▓██·▓█·▒▒··▒▓··█·█·▒·▓·░····█····░···█·█·▓▒▒··▒▒···
    Less ·░▒▓█ More

Jan  ██████████████████▏   28      18h
Feb  ████████████▉         18  10h 10m
...
Dec  ███▉                   6   2h 50m

Mon  ██████████▊           54%
Tue  ███████████▌          58%
...
Sun  █████████▎            46%
```

#### Trends

```bash
obser habits trend [--by month|year] [--metric rate|days|total|avg] [--year YYYY]
```

Shows a grid with one row per habit and one column per month of a year (`--by month`, the default, for the current year or `--year`) or per year since the first journal entry (`--by year`, which doesn't take `--year`). Each cell shows the completion rate by default, or the days done, total or average amount with `--metric`. Months after today and before a habit started are blank, and `total`/`avg` show `-` for habits without amounts. The `Δ` column compares the last complete period with the one before it: `↑` or `↓` when it changed by more than 10%, `→` otherwise. With colors on, each cell is shaded relative to the highest value in its row.

```bash
obser habits trend
obser habits trend --year 2025 --metric total
obser habits trend --by year
```

```
HABIT         2023  2024  2025  2026  Δ
calisthenics   17%   21%   21%   16%  →
chain                65%   81%   75%  ↑
guitar         21%   33%   46%   27%  ↑
read           11%    5%   47%   39%  ↑
```
