package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Database handles PostgreSQL operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	// Get connection string from environment
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("[Database] Connected to PostgreSQL")

	return &Database{db: db}, nil
}

// InitSchema ensures basic connectivity - migrations handled by Atlas
func (d *Database) InitSchema() error {
	log.Println("[Database] Schema management handled by Atlas migrations")
	return nil
}

// SaveProject saves a project and its cells to the database
func (d *Database) SaveProject(project *Project) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or update project
	_, err = tx.Exec(`
		INSERT INTO projects (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`, project.ID, project.Name, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	// Delete existing cells for this project
	_, err = tx.Exec("DELETE FROM cells WHERE project_id = $1", project.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old cells: %w", err)
	}

	// Insert all cells with their index position
	for i, cell := range project.Cells {
		cellData, err := json.Marshal(cell)
		if err != nil {
			return fmt.Errorf("failed to marshal cell: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO cells (id, project_id, index_position, type, data, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, cell.GetID(), project.ID, i, cell.GetType(), cellData, cell.GetCreatedAt(), cell.GetUpdatedAt())
		if err != nil {
			return fmt.Errorf("failed to save cell: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[Database] Saved project: %s with %d cells", project.ID, len(project.Cells))
	return nil
}

// LoadProject loads a project and its cells from the database
func (d *Database) LoadProject(projectID string) (*Project, error) {
	// Load project metadata
	var project Project
	err := d.db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM projects
		WHERE id = $1
	`, projectID).Scan(&project.ID, &project.Name, &project.CreatedAt, &project.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Project doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Load cells ordered by index_position
	rows, err := d.db.Query(`
		SELECT id, type, data, created_at, updated_at
		FROM cells
		WHERE project_id = $1
		ORDER BY index_position ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cells: %w", err)
	}
	defer rows.Close()

	project.Cells = []Cell{}
	for rows.Next() {
		var cellID, cellType string
		var cellData []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&cellID, &cellType, &cellData, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cell: %w", err)
		}

		// Unmarshal cell based on type
		var cell Cell
		switch cellType {
		case "equation":
			var eq EquationCell
			if err := json.Unmarshal(cellData, &eq); err != nil {
				return nil, fmt.Errorf("failed to unmarshal equation cell: %w", err)
			}
			cell = eq
		case "folder":
			var folder FolderCell
			if err := json.Unmarshal(cellData, &folder); err != nil {
				return nil, fmt.Errorf("failed to unmarshal folder cell: %w", err)
			}
			cell = folder
		case "note":
			var note NoteCell
			if err := json.Unmarshal(cellData, &note); err != nil {
				return nil, fmt.Errorf("failed to unmarshal note cell: %w", err)
			}
			cell = note
		default:
			log.Printf("[Database] Unknown cell type: %s", cellType)
			continue
		}

		project.Cells = append(project.Cells, cell)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cells: %w", err)
	}

	log.Printf("[Database] Loaded project: %s with %d cells", project.ID, len(project.Cells))
	return &project, nil
}

// ProjectExists checks if a project exists in the database
func (d *Database) ProjectExists(projectID string) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}

// DeleteProject removes a project and its cells from the database
func (d *Database) DeleteProject(projectID string) error {
	_, err := d.db.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	log.Printf("[Database] Deleted project: %s", projectID)
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
