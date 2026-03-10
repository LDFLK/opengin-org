<!-- Banner/logo placeholder -->

# OpenGIN-Org

[![License: Apache 2.0](https://img.shields.io/badge/Code-Apache%202.0-blue.svg)](LICENSE)
[![License: CC BY-NC-SA 4.0](https://img.shields.io/badge/Data-CC%20BY--NC--SA%204.0-77DD77.svg)](https://creativecommons.org/licenses/by-nc-sa/4.0/)

A command-line tool for managing and processing organisational structure data, specifically designed for tracking government ministries, departments, and personnel appointments in Sri Lanka.

## Features

| Feature | Description |
|---|---|
| Data Processing | Processes transaction files from a specified directory to build the organisational structure |
| Entity Management | Creates and manages government entities, ministry appointments, department structures, and personnel assignments |
| Relationship Tracking | Maintains hierarchical relationships between entities, tracks appointment dates and durations, records historical changes |
| API Integration | Separate endpoints for updates and queries via a RESTful API interface |
| Process Types | Organisation mode for minister/department entities, People mode for citizen entities, Document mode for gazette documents |

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Access to the required API endpoints
- Transaction data in the specified CSV format

### Build

```bash
go build -o orgchart cmd/main.go
```

### Basic Usage

```bash
# Show help and usage information
./orgchart --help

# Process organisation data
./orgchart -data /path/to/data/directory

# Process people data
./orgchart -data /path/to/data/directory -type people

# Initialize database and process organisation data
./orgchart -data /path/to/data/directory -init
```

For detailed CLI options, data structure conventions, API endpoints, and development guides, see [DEVELOPMENT.md](DEVELOPMENT.md).

## Contributing

We welcome contributions from the community. Please read our [Contributing Guide](CONTRIBUTING.md) for details on how to get started, branching strategy, commit conventions, and the pull request process.

## Code of Conduct

This project follows the Lanka Data Foundation Code of Conduct. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before participating.

## Security

To report security vulnerabilities, please follow the process outlined in [SECURITY.md](SECURITY.md). Do not open public issues for security concerns.

## License

This project uses dual licensing:

- **Code**: Licensed under the [Apache License 2.0](LICENSE)
- **Data** (files under `data/`): Licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/)

See [LICENSE](LICENSE) for the full Apache 2.0 license text.

## References

- [Lanka Data Foundation](https://datafoundation.lk)
- [Neo4j Graph Database](https://neo4j.com)
