# Document Linking Program

This program creates relationships between document entities based on CSV data files.

## Overview

The `link_documents.go` program reads CSV files containing document relationship data and creates the corresponding relationships in the entity system. Each CSV file contains information about document relationships with the following structure:

- `parent`: Name of the parent document
- `child`: Name of the child document  
- `relationship`: Type of relationship (e.g., "AMENDS", "REFERS_TO")
- `start_date`: When the relationship starts (YYYY-MM-DD format)

## CSV File Format

The CSV files should have the following structure:

```csv
parent,child,relationship,start_date
2403-53,2403-39,AMENDS,2024-09-27
2403-53,2403-38,AMENDS,2024-09-25
2411-09,2403-53,AMENDS,2024-11-18
```

## How It Works

1. **CSV Reading**: The program reads all CSV files from the `docs_linking_data/` directory
2. **Document Search**: For each relationship, it searches for both parent and child documents by name (MajorKind: Document)
3. **Relationship Creation**: Creates the specified relationship between the documents with the given start date
4. **Error Handling**: Provides detailed logging for successful operations and errors

## Usage

### Prerequisites

- The API server must be running on `http://localhost:8080`
- Document entities must already exist in the system
- CSV files must be placed in the `docs_linking_data/` directory

### Running the Program

```bash
cd scripts
go run link_documents.go
```

### Configuration

The program uses the following default configuration:
- Update URL: `http://localhost:8080/api/v1/entities`
- Query URL: `http://localhost:8080/api/v1/query`

To modify these URLs, edit the constants in the `main()` function.

## File Structure

```
scripts/
├── link_documents.go          # Main program
├── docs_linking_data/         # CSV files directory
│   ├── docs_linking_anura.csv
│   ├── docs_linking_gota.csv
│   └── docs_linking_ranil.csv
└── README_document_linking.md # This file
```

## Error Handling

The program includes comprehensive error handling for:

- Missing or malformed CSV files
- Document entities not found in the system
- Invalid date formats
- API communication errors
- Duplicate or conflicting relationships

## Output

The program provides detailed console output including:

- Progress information for each CSV file processed
- Success/failure status for each relationship creation
- Summary statistics at the end
- Error messages with context for debugging

## Example Output

```
Found 3 CSV files to process

=== Processing docs_linking_data/docs_linking_anura.csv ===
Processing 6 document links from docs_linking_data/docs_linking_anura.csv
Processing link 1/6: 2403-53 -> 2403-39 (AMENDS)
Successfully created relationship: 2403-53 (doc_001) -> 2403-39 (doc_002) [AMENDS]
...
Completed processing docs_linking_data/docs_linking_anura.csv: 6 successful, 0 errors

=== Summary ===
Successfully processed: 3 files
Failed to process: 0 files
Total files: 3
```

## Notes

- All document entities must have MajorKind "Document"
- The program searches for documents by name, so names must match exactly
- Relationships are created with the specified start date and no end date
- The program processes all CSV files in the directory automatically
