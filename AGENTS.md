# AGENTS.md

## Repository Overview

### Project Description
**excel-template-mapper** is a Proof of Concept (PoC) that creates a foundation for converting spreadsheet formats (paper-like Excel files) into JSON (structured data). The primary goal is to convert Japanese Excel reports with irregular layouts into searchable, inspectable JSON with Japanese keys.

**Main purpose:**
- Convert paper-style Excel spreadsheets → JSON with Japanese keys
- Enable grep/regex-based search and inspection of report data
- Template-based mapping system for reusable conversion logic
- Database-backed template storage (MySQL) for sharing across workflows

**Key technologies:**
- Go 1.24+
- GraphQL (gqlgen) for API
- MySQL 8.0 for template/storage
- Excelize v2 for Excel file reading
- Docker Compose for local development

---

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     User Interface Layer                    │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │  CLI (etm)   │      │ GraphQL API │                     │
│  │  cmd/etm     │      │ cmd/api      │                     │
│  └──────┬───────┘      └──────┬───────┘                     │
└─────────┼─────────────────────┼─────────────────────────────┘
          │                     │
          ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Business Logic Layer                     │
│  ┌──────────────────┐        ┌─────────────────────────┐    │
│  │   Excel Reader   │        │      Processor          │    │
│  │ internal/excel   │        │   internal/service      │    │
│  └────────┬─────────┘        └────────────┬────────────┘    │
└───────────┼───────────────────────────────┼─────────────────┘
            │                               │
            ▼                               ▼
┌─────────────────────────────────────────────────────────────┐
│                     Data Access Layer                       │
│  ┌──────────────────────┐      ┌────────────────────────┐   │
│  │    MySQL Store       │      │    DB Connection       │   │
│  │ internal/store/mysql │      │ internal/dbconn        │   │
│  └──────────┬───────────┘      └──────────┬─────────────┘   │
└─────────────┼─────────────────────────────┼─────────────────┘
              │                             │
              ▼                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                             │
│  ┌──────────────────┐         ┌─────────────────────────┐   │
│  │  MySQL DB        │         │    Excel Files          │   │
│  │ mapping_templates│         │   (input files)         │   │
│  │ mapping_rules    │         │                         │   │
│  └──────────────────┘         └─────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**Data Flow:**
1. User provides Excel file and template name (or creates new template)
2. Template defines: sheet, header row, data start row, column mappings
3. Reader parses Excel → Table structure with headers + rows
4. Processor applies mapping rules → converts to JSON output
5. JSON uses Japanese labels as keys (default) for searchability

**Key Components:**
- **Template**: Defines the spreadsheet structure (sheet name, header position, etc.)
- **Mapping Rule**: Maps source columns to output fields with transforms
- **Processor**: Executes conversion pipeline (reader → transformer → emitter)

---

### Directory Structure

```
.
├── cmd/
│   ├── api/                # GraphQL API server entry point
│   │   └── main.go        # Runs GraphQL playground on :8080
│   └── etm/               # CLI tool for Excel→JSON conversion
│       └── main.go        # Command: `etm --template <name> <file.xlsx>`
├── db/
│   ├── migrations/        # Database migration scripts
│   │   ├── 001_init.up.sql    # Create tables
│   │   └── 001_init.down.sql  # Drop tables
│   └── mysql/conf.d/      # MySQL configuration overrides
├── graph/
│   ├── generated.go       # Auto-generated GraphQL execution code
│   ├── loader/loader.go   # Dataloader for batching (if needed)
│   ├── model/models_gen.go    # Generated GraphQL models
│   ├── schema.graphqls    # GraphQL schema definition
│   └── schema.resolvers.go  # Resolver implementations
├── internal/
│   ├── dbconn/            # Database connection utilities
│   │   └── dbconn.go      # Load MySQL config from env, Ping, ListTables
│   ├── excel/             # Excel reading logic
│   │   ├── reader.go      # ReadTable function for header+rows parsing
│   │   └── Book1.xlsx     # Sample Excel file for testing
│   ├── model/             # Core domain models (not GraphQL-specific)
│   │   ├── template.go    # Template struct
│   │   └── rule.go        # Rule struct
│   ├── service/           # Business logic layer
│   │   └── processor.go   # ProcessExcelOnly for direct conversion
│   └── store/mysql/       # MySQL data access layer
│       └── template_repo.go  # GetTemplateByName, ListRulesByTemplateID
├── .continue/             # AI assistant configuration
│   └── rules/
│       └── review.md      # Custom code review command
├── server.go              # Alternative GraphQL server entry point
├── docker-compose.yml     # MySQL service definition
├── gqlgen.yml             # GraphQL code generation config
├── Makefile               # Convenience commands for DB operations
├── .env                   # Environment variables (not tracked)
└── README.md              # Project documentation
```

**Entry Points:**
- `cmd/api/main.go` → GraphQL server on port 8080 with playground
- `cmd/etm/main.go` → CLI tool for batch Excel→JSON conversion

---

### Development Workflow

#### Setup & Run

1. **Start MySQL via Docker Compose:**
   ```bash
   make db-up
   ```

2. **Run database migrations:**
   ```bash
   make migrate-up
   ```

3. **Create `.env` file** (based on `internal/dbconn/dbconn.go`):
   ```
   SQL_USER=app
   SQL_PASSWORD=app
   SQL_ADDR=localhost:3306
   SQL_DBNAME=excel_template_mapper
   # Optional: PORT=8080
   ```

4. **Run GraphQL API server:**
   ```bash
   go run cmd/api/main.go
   ```
   Then open http://localhost:8080 in browser for GraphQL playground.

5. **Use CLI tool:**
   ```bash
   # Convert Excel to JSON using a saved template
   etm --template demo_v1 path/to/file.xlsx
   
   # Enable debug logs
   etm --template demo_v1 --debug path/to/file.xlsx
   ```

#### Testing Approach

- **No test suite yet** - PoC stage
- Manual testing via:
  - GraphQL playground for template/RULE management
  - CLI tool with sample Excel files (`internal/excel/Book1.xlsx`)
  - Database inspection: `make mysql` then run SQL queries

#### Lint & Format Commands

```bash
# Format code
gofmt -w ./...

# Build
go build ./...

# Run all tests (none yet)
go test ./...
```

---

### Key Concepts

**Template System:**
- Templates define spreadsheet structure in MySQL
- Reusable across multiple Excel files
- Supports both header-based (table) and cell-based (form) mappings

**Mapping Rules:**
- `HEADER` type: Map by column header text (flexible column order)
- `CELL` type: Map by specific cell address (forms with fixed layout)
- Optional transforms: `trim`, `date:2006-01-02`
- Required field validation

**Output Format:**
- Default: Japanese keys in JSON (for grep/regex search)
- Future: `canonical_key` option for English API keys
- Each row = one record, columns = fields

---

### Environment Variables

Required (per `internal/dbconn/dbconn.go`):
- `SQL_USER`
- `SQL_PASSWORD`
- `SQL_ADDR` (e.g., `localhost:3306`)
- `SQL_DBNAME`

Optional:
- `PORT` (default: 8080)
- `SQL_NET` (default: `tcp`, also accepts typo `SQL_NST`)

---

### Database Schema

**Table: `mapping_templates`**
| Column | Type | Description |
|--------|------|-------------|
| id | VARCHAR(36) | UUID template identifier |
| name | VARCHAR(255) | Unique template name (e.g., "reporting_v1") |
| target | VARCHAR(255) | Logical target name |
| sheet_name | VARCHAR(255) | Target sheet (nullable) |
| header_row | INT | Header row number (1-based) |
| data_start_row | INT | First data row (1-based) |
| notes | VARCHAR(1024) | Optional documentation |
| created_at/updated_at | DATETIME(6) | Timestamps |

**Table: `mapping_rules`**
| Column | Type | Description |
|--------|------|-------------|
| id | VARCHAR(36) | UUID rule identifier |
| template_id | VARCHAR(36) | FK to templates |
| source_type | VARCHAR(16) | "HEADER" or "CELL" |
| source_key | VARCHAR(255) | Header text or cell address (e.g., "E17") |
| target_id | VARCHAR(36) | Internal item ID |
| target_label | VARCHAR(255) | **Output JSON key** (Japanese default) |
| canonical_key | VARCHAR(255) | English API key (optional, future use) |
| transform | VARCHAR(255) | e.g., "trim", "date:..." |
| required | TINYINT(1) | 0/1 for validation |
| priority | INT | For duplicate handling |
| evidence | TEXT | AI reasoning or human notes |
| created_at/updated_at | DATETIME(6) | Timestamps |

---

### Current Status & Non-Goals

**Implemented:**
- Excel reading with flexible header positions
- Template storage in MySQL
- Header-based mapping rules with transforms
- JSON output with Japanese keys
- CLI tool for batch conversion

**Not Implemented (Future Work):**
- GraphQL mutations (create/update templates/rules)
- Cell-based (CELL type) rule processing
- AI-based layout inference
- Web UI
- Complex transform DSL
- Multiple data regions in one Excel
- Validation/error reporting improvements

---

### Quick Start Examples

1. **Create a template manually in MySQL:**
   ```sql
   INSERT INTO mapping_templates VALUES (
     UUID(), 'demo_v1', 'CustomerReport', NULL, 1, 2, '', NOW(), NOW()
   );
   
   INSERT INTO mapping_rules VALUES (
     UUID(), '<template_id>', 'HEADER', 'カスタマID',
     UUID(), 'カスタマID', NULL, NULL, 1, 1, '', NOW(), NOW()
   );
   ```

2. **Convert Excel using template:**
   ```bash
   etm --template demo_v1 internal/excel/Book1.xlsx
   ```

3. **Access GraphQL playground:**
   ```bash
   go run cmd/api/main.go
   # Visit http://localhost:8080
   ```

---

### Troubleshooting

- **"missing env" error**: Ensure `.env` has all required variables
- **Connection refused**: Run `make db-up` and check MySQL is listening on 3306
- **Template not found**: Verify template name in database: `SELECT name FROM mapping_templates;`
- **Required header missing**: Check Excel headers match rule `source_key` exactly (trimmed)
