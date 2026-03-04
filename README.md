# organisation Chart Tool

A command-line tool for managing and processing organisational structure data, specifically designed for tracking government ministries, departments, and personnel appointments.

## Overview

This tool processes transaction data to build and maintain an organisational chart that tracks:
- Government ministries and their relationships
- Department structures under ministries
- Personnel appointments and their changes over time
- Historical tracking of organisational changes

## Features

- **Data Processing**: Processes transaction files from a specified directory to build the organisational structure
- **Entity Management**: 
  - Creates and manages government entities
  - Tracks ministry appointments
  - Manages department structures
  - Handles personnel assignments
- **Relationship Tracking**:
  - Maintains hierarchical relationships between entities
  - Tracks appointment dates and durations
  - Records historical changes in organisational structure
- **API Integration**:
  - Separate endpoints for updates and queries
  - RESTful API interface for data operations
- **Process Types**:
  - Organisation mode: Processes minister and department entities
  - People mode: Processes citizen entities

## Building the Tool

To build the executable from the base directory:

```bash
go build -o orgchart cmd/main.go
```

This will create an executable named `orgchart` in the current directory.

## Usage

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

# Process secretary operations (cascade logic for ministry renames)
./orgchart -process-secretary-operations
```

### Command Line Options

- `-data`: (Required for data processing) Path to the data directory containing transactions
- `-init`: (Optional) Initialize the database with government node
- `-type`: (Optional) Type of data to process: 'organisation' or 'people' (default: organisation)
- `-update_endpoint`: (Optional) Endpoint for the Update API (default: "http://localhost:8080/entities")
- `-query_endpoint`: (Optional) Endpoint for the Query API (default: "http://localhost:8081/v1/entities")
- `-process-secretary-operations`: (Optional) Process secretary cascade operations for ministry renames

### Process Types

The tool supports two modes of operation:

1. **organisation Mode** (default):
   - Processes minister and department entities
   - Tracks organisational structure
   - Manages hierarchical relationships

2. **People Mode**:
   - Processes citizen entities
   - Tracks personnel appointments
   - Manages individual relationships

## Data Structure

The tool processes transaction files that define:
1. **Ministries**: Government ministries and their appointments
2. **Departments**: organisational units under ministries
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

**Note**: If a CSV file contains a `president` column with a value, that value will be used instead of the directory-derived name. If the `president` column is empty or missing, the system falls back to using the president name from the directory structure. (This is useful when you are creating csv files for moving between presidents and need to specify two different president names or a president name different from the current president's)

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
1. **Update API**: Handles all write operations (default: http://localhost:8080/entities)
2. **Query API**: Handles all read operations (default: http://localhost:8081/v1/entities)

## Requirements

- Go 1.x or higher
- Access to the required API endpoints
- Transaction data in the specified format
- CSV files following the required naming convention

## Insert Data

### Insert Minister Department

```bash
./orgchart -data $(pwd)/data/orgchart/akd/2024-09-27/ -init true
./orgchart -data $(pwd)/data/people/akd/2024-09-25/ -type person
```

## Secretary Operations

### What It Does

When a ministry is renamed (e.g., "Minister of Roads and Highways" → "Minister of Highways"), the tool:

1. **Detects Ministry Renames**: Finds all `RENAMED_TO` relationships
2. **Processes Old Ministry Secretaries**: 
   - Applies termination chain (each secretary ends when next starts)
   - Terminates active secretary on rename date
3. **Moves Secretaries**: Moves secretary to new ministry if terminated on rename date
4. **Processes New Ministry Secretaries**: Creates continuous service records

### Running Secretary Operations

```bash
# Process all secretary operations (must be run after data is loaded)
./orgchart -process-secretary-operations
```

**Note**: This command should be run after loading all ministry and secretary data into the database.

### Helper Scripts

#### Load All Secretaries

A convenience script to load all secretary data at once:

```bash
# Load all secretary data for all presidents
./load_all_secretaries.sh
```bash
./load_all_secretaries.sh
./orgchart -process-secretary-operations
```

### Testing Secretary Operations

The project includes comprehensive test suites for secretary operations:

```bash
# Run all secretary operation tests (unit + integration)
cd tests
go test -v secretary_operations_test.go

# Run only unit tests 
go test -v -short secretary_operations_test.go

# Run real-world validation tests
go test -v Real_secretary_operations_test.go


## Development

The project structure:
```
.
├── cmd/
│   └── main.go         # Main application entry point
├── api/                # API client and operations
├── models/             # Data models and structures
└── tests/              # Test files
```

## License

[Add your license information here]


### To take a dump and reupload it to your local instance

```bash
docker run --rm \
--volume=/var/lib/docker/volumes/neo4j_data/_data:/data \
--volume=/Users/zaeema/Documents/neo4j_dump/orgchart:/backups \
neo4j/neo4j-admin:latest \
neo4j-admin database dump neo4j --to-path=/backups
```

```bash
docker run --interactive --tty --rm \
  --volume /var/lib/docker/volumes/neo4j_data/_data:/data \
  --volume /Users/zaeema/Documents/neo4j_dump/orgchart-gota-2025-09-11:/backups \
  neo4j/neo4j-admin:5 \
  neo4j-admin database load neo4j --from-path=/backups --overwrite-destination=true
```

Gota & ranil
```bash
docker run --interactive --tty --rm \
  --volume /var/lib/docker/volumes/neo4j_data/_data:/data \
  --volume /Users/zaeema/Documents/neo4j_dump/orgchart-gota-ranil:/backups \
  neo4j/neo4j-admin:5 \
  neo4j-admin database load neo4j --from-path=/backups --overwrite-destination=true
```



