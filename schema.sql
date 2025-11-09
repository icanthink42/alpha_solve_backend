-- Alpha Solve Database Schema

CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cells (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    index_position INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cells_project_id ON cells(project_id);
CREATE INDEX IF NOT EXISTS idx_cells_project_index ON cells(project_id, index_position);

