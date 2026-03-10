# Development Guide

This document contains detailed developer and operational documentation for the Organisation Chart Tool. For a project overview, see [README.md](README.md).

## CLI Usage

The tool can be run with various options:

```bash
# Show help and usage information
./orgchart --help

# Process organisation data with default settings
./orgchart -data /path/to/data/directory

# Process people data
./orgchart -data /path/to/data/directory -type people

# Process document data
./orgchart -data /path/to/data/directory -type document

# Initialize database and process organisation data
./orgchart -data /path/to/data/directory -init

# Use custom API endpoints
./orgchart -data /path/to/data/directory -update_endpoint http://custom:8080/entities -query_endpoint http://custom:8081/v1/entities
```

### Command Line Options

| Option | Required | Description | Default |
|---|---|---|---|
| `-data` | Yes | Path to the data directory containing transactions | — |
| `-init` | No | Initialize the database with government node | `false` |
| `-type` | No | Type of data to process: `organisation`, `people`, or `document` | `organisation` |
| `-update_endpoint` | No | Endpoint for the Update API | `http://localhost:8080/entities` |
| `-query_endpoint` | No | Endpoint for the Query API | `http://localhost:8081/v1/entities` |

## Process Types

The tool supports three modes of operation:

### Organisation Mode (default)

- Processes minister and department entities
- Tracks organisational structure
- Manages hierarchical relationships

### People Mode

- Processes citizen entities
- Tracks personnel appointments
- Manages individual relationships

### Document Mode

- Processes gazette document entities

## Data Structure

The tool processes transaction files that define:

1. **Ministries**: Government ministries and their appointments
2. **Departments**: Organisational units under ministries
3. **Personnel**: People appointed to various positions
4. **Relationships**: Hierarchical and appointment relationships between entities

### Directory Structure and President Name Extraction

The tool automatically extracts the president's name from the directory structure. The directory path must follow this pattern:

```
data/
├── orgchart/
│   └── PresidentName/
│       └── Date/
│           └── transaction_files.csv
├── people/
│   └── PresidentName/
│       └── Date/
│           └── transaction_files.csv
└── documents/
    └── PresidentName/
        └── Date/
            └── transaction_files.csv
```

**Important**: The directory name immediately after `orgchart/`, `people/`, or `documents/` is used as the president's name for all transactions in that directory tree.

**Example Directory Structure**:

```
data/orgchart/Ranil Wickremesinghe/2024-09-27/2403_53_ADD.csv
data/people/Anura Kumara Dissanayake/2024-09-23/2403-03_ADD.csv
data/documents/Maithripala Sirisena/2018-11-01/2095_17_ADD.csv
```

**Note**: If a CSV file contains a `president` column with a value, that value will be used instead of the directory-derived name. If the `president` column is empty or missing, the system falls back to using the president name from the directory structure. (This is useful when you are creating CSV files for moving between presidents and need to specify two different president names or a president name different from the current president's.)

### Transaction File Naming Convention

Transaction files must follow a specific naming convention:

- Files must contain `_ADD` in their name to be recognized as ADD transactions
- The `_ADD` can be at the end of the filename or preceded by a prefix
- Valid examples:
  - `ADD.csv`
  - `2403-38_ADD.csv`
  - `Xpr_ADD.csv`
  - `2024_03_ADD.csv`

The tool will process all CSV files in the specified directory that match this naming pattern.

## API Endpoints

The tool uses two main API endpoints:

1. **Update API**: Handles all write operations (default: `http://localhost:8080/entities`)
2. **Query API**: Handles all read operations (default: `http://localhost:8081/v1/entities`)

## Insert Data

### Insert Minister Department

```bash
./orgchart -data $(pwd)/data/orgchart/akd/2024-09-27/ -init true
./orgchart -data $(pwd)/data/people/akd/2024-09-25/ -type person
```

## Neo4j Database Backup and Restore

### Taking a Database Dump

```bash
docker run --rm \
  --volume=/var/lib/docker/volumes/neo4j_data/_data:/data \
  --volume=/path/to/backups:/backups \
  neo4j/neo4j-admin:latest \
  neo4j-admin database dump neo4j --to-path=/backups
```

### Restoring a Database Dump

```bash
docker run --interactive --tty --rm \
  --volume /var/lib/docker/volumes/neo4j_data/_data:/data \
  --volume /path/to/backups:/backups \
  neo4j/neo4j-admin:5 \
  neo4j-admin database load neo4j --from-path=/backups --overwrite-destination=true
```

## Project Structure

```
.
├── cmd/
│   └── main.go         # Main application entry point
├── api/                # API client and operations
├── models/             # Data models and structures
├── data/               # Transaction data files
└── tests/              # Test files
```

## Testing

Run all tests:

```bash
go test ./tests
```

Run an individual test:

```bash
go test -v -run ^TestCreateMinisters$
```

If you change the API code (i.e. Nexoan) but not the test code, run the following to execute the tests without caching the previous test results:

```bash
go test ./tests -count=1
```
