# RepClient

**RepClient** is a fast, flexible simple CLI tool for querying and streaming DNS records (A, AAAA, NS, MX, TXT, CNAME) from rep API with support for config files, multithreading, and output to multiple formats.

## Demo

[![asciicast](https://asciinema.org/a/MaSyT7OJCYKsFgeS5osoeG533.svg)](https://asciinema.org/a/MaSyT7OJCYKsFgeS5osoeG533)

## Features

- Load DNS records from list files or streams
- Process A and AAAA DNS records with typed handling
- Read input from file with auto type-detection
- Utility functions for saving output in `.json`, `.csv`, `.txt`, or `.ndjson` formats
- Flexible command-line argument parsing
- Configurable via TOML
- Verbose and structured logging with `logrus`
- and many more. you can check yourself

## Installation

### 📁 From Release

Download the latest precompiled binary from the [Releases page](https://github.com/Doom-z/RepClient/releases).


### 🔧 From Source

```bash
git clone https://github.com/Doom-z/RepClient.git
cd RepClient
cp config.example.toml config.toml
go mod tidy
go build -o repclient ./cmd
```



> Choose the appropriate binary for your OS and architecture, then give it executable permissions if needed:
>
> ```bash
> chmod +x repclient
> ```


### Options

| Flag                    | Description                                                  | Default       |
| ----------------------- | ------------------------------------------------------------ | ------------- |
| `--trial`          | Free Version Query supported all dns except ipv6 with limited results ( upto 1k ) | `false` if apikey is `@repproject` forced set to `true`
| `--ipv4`, `-i`          | Query A record for given IPv4 address                        |               |
| `--ipv6`                | Query AAAA record for given IPv6 address (requires `--full`) |               |
| `--ns`, `-s`            | Query NS record                                              |               |
| `--cname`, `-n`         | Query CNAME record                                           |               |
| `--txt`, `-t`           | Query TXT record                                             |               |
| `--mx`, `-m`            | Query MX record                                              |               |
| `--list-file`, `-l`     | Input file with multiple entries                             |               |
| `--full`, `-f`          | Use full mode (for A/AAAA record streaming) use this if you want to use this your output format must be `ndjson`                 | `false`       |
| `--max-total-output-ip`, `-m` | Maximum records to fetch per IP                              | `100`         |
| `--page-size`, `-p`     | Page size for pagination                                     | `100`         |
| `--output`, `-o`        | Write results to output file                                 | `false`       |
| `--threads`, `-t`       | Number of threads to use when reading list files             | `1`           |
| `--verbose`, `-v`       | Enable verbose logging                                       | `false`       |
| `--config`, `-c`        | Path to TOML config file                                     | `config.toml` |


---

## 🧩 Example Workflows

### Query a Single IP

```bash
./repclient -i 8.8.8.8 -o
```

### Query using custom name file config

```bash
./repclient --ipv6 2606:4700:4700::1111 -c myconfig.toml
```

### Query in Full Mode (A or AAAA) with output enabled

```bash
./repclient --ipv6 2606:4700:4700::1111 --full -o
```

### Query Using a List File with output enabled and 5 threads

```bash
./repclient -l targets.txt -o --threads 5
```

---


## 🙌 Contributing
PRs welcome! Please open an issue for discussion before submitting breaking changes.
