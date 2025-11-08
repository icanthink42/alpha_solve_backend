package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// ProjectManager manages project state
type ProjectManager struct {
	projects map[string]*Project
	mu       sync.RWMutex
}

// NewProjectManager creates a new ProjectManager
func NewProjectManager() *ProjectManager {
	return &ProjectManager{
		projects: make(map[string]*Project),
	}
}

// GetProject retrieves a project by ID
func (pm *ProjectManager) GetProject(projectID string) *Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.projects[projectID]
}

// SetProject stores or updates a project
func (pm *ProjectManager) SetProject(project *Project) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.projects[project.ID] = project
}

// UpdateCell updates a specific cell in a project, or adds it if it doesn't exist
func (pm *ProjectManager) UpdateCell(projectID string, cellID string, cell Cell) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	project, exists := pm.projects[projectID]
	if !exists {
		return false
	}

	// Try to update the cell in the project's cell array
	updated := updateCellRecursive(project.Cells, cellID, cell)

	// If cell wasn't found, add it to the project
	if !updated {
		project.Cells = append(project.Cells, cell)
		log.Printf("[ProjectManager] Added new cell: %s to project: %s", cellID, projectID)
	}

	project.UpdatedAt = time.Now()
	return true
}

// updateCellRecursive recursively searches and updates a cell
func updateCellRecursive(cells []Cell, cellID string, newCell Cell) bool {
	for i, cell := range cells {
		if cell.GetID() == cellID {
			cells[i] = newCell
			return true
		}

		// If it's a folder, search recursively
		if folderCell, ok := cell.(FolderCell); ok {
			if updateCellRecursive(folderCell.Cells, cellID, newCell) {
				return true
			}
		}
	}
	return false
}

// MoveCell moves a cell from one index to another
func (pm *ProjectManager) MoveCell(projectID string, cellID string, fromIndex int, toIndex int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	project, exists := pm.projects[projectID]
	if !exists {
		return false
	}

	// Validate indices
	if fromIndex < 0 || fromIndex >= len(project.Cells) || toIndex < 0 || toIndex >= len(project.Cells) {
		log.Printf("[ProjectManager] Invalid indices: from=%d, to=%d, len=%d", fromIndex, toIndex, len(project.Cells))
		return false
	}

	// Move the cell
	cell := project.Cells[fromIndex]
	project.Cells = append(project.Cells[:fromIndex], project.Cells[fromIndex+1:]...)

	// Insert at new position
	if toIndex > fromIndex {
		toIndex--
	}
	project.Cells = append(project.Cells[:toIndex], append([]Cell{cell}, project.Cells[toIndex:]...)...)

	project.UpdatedAt = time.Now()
	return true
}

// CreateCell creates a new cell at a specific position (optionally inside a parent folder)
func (pm *ProjectManager) CreateCell(projectID string, cell Cell, parentCellID *string, index int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	project, exists := pm.projects[projectID]
	if !exists {
		return false
	}

	// If no parent, insert at root level
	if parentCellID == nil {
		// Validate index
		if index < 0 || index > len(project.Cells) {
			log.Printf("[ProjectManager] Invalid index: %d, len=%d", index, len(project.Cells))
			return false
		}

		// Insert at index
		project.Cells = append(project.Cells[:index], append([]Cell{cell}, project.Cells[index:]...)...)
		project.UpdatedAt = time.Now()
		return true
	}

	// Insert inside a parent folder
	success := createCellInFolder(project.Cells, cell, *parentCellID, index)
	if success {
		project.UpdatedAt = time.Now()
	}
	return success
}

// createCellInFolder recursively finds a folder and inserts a cell
func createCellInFolder(cells []Cell, newCell Cell, parentCellID string, index int) bool {
	for i, cell := range cells {
		if cell.GetID() == parentCellID {
			// Found the parent folder
			if folderCell, ok := cell.(FolderCell); ok {
				// Validate index
				if index < 0 || index > len(folderCell.Cells) {
					log.Printf("[ProjectManager] Invalid index in folder: %d, len=%d", index, len(folderCell.Cells))
					return false
				}

				// Insert at index
				folderCell.Cells = append(folderCell.Cells[:index], append([]Cell{newCell}, folderCell.Cells[index:]...)...)
				cells[i] = folderCell
				return true
			}
			return false
		}

		// Recursively search in nested folders
		if folderCell, ok := cell.(FolderCell); ok {
			if createCellInFolder(folderCell.Cells, newCell, parentCellID, index) {
				cells[i] = folderCell
				return true
			}
		}
	}
	return false
}

// DeleteCell deletes a cell by ID (optionally from inside a parent folder)
func (pm *ProjectManager) DeleteCell(projectID string, cellID string, parentCellID *string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	project, exists := pm.projects[projectID]
	if !exists {
		return false
	}

	// If no parent, delete from root level
	if parentCellID == nil {
		for i, cell := range project.Cells {
			if cell.GetID() == cellID {
				project.Cells = append(project.Cells[:i], project.Cells[i+1:]...)
				project.UpdatedAt = time.Now()
				return true
			}
		}
		return false
	}

	// Delete from inside a parent folder
	success := deleteCellFromFolder(project.Cells, cellID, *parentCellID)
	if success {
		project.UpdatedAt = time.Now()
	}
	return success
}

// deleteCellFromFolder recursively finds a folder and deletes a cell
func deleteCellFromFolder(cells []Cell, cellID string, parentCellID string) bool {
	for i, cell := range cells {
		if cell.GetID() == parentCellID {
			// Found the parent folder
			if folderCell, ok := cell.(FolderCell); ok {
				for j, childCell := range folderCell.Cells {
					if childCell.GetID() == cellID {
						folderCell.Cells = append(folderCell.Cells[:j], folderCell.Cells[j+1:]...)
						cells[i] = folderCell
						return true
					}
				}
			}
			return false
		}

		// Recursively search in nested folders
		if folderCell, ok := cell.(FolderCell); ok {
			if deleteCellFromFolder(folderCell.Cells, cellID, parentCellID) {
				cells[i] = folderCell
				return true
			}
		}
	}
	return false
}

// HandlePacket processes incoming packets and returns a response
// Returns (response []byte, shouldBroadcast bool)
func HandlePacket(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	// First, parse to get the packet type
	var basePacket struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &basePacket); err != nil {
		log.Printf("[Handler] Failed to parse packet type: %v", err)
		return nil, false
	}

	log.Printf("[Handler] Processing packet type: %s for project: %s", basePacket.Type, projectID)

	switch basePacket.Type {
	case "ProjectSync":
		return handleProjectSync(pm, projectID, data)

	case "CellUpdate":
		return handleCellUpdate(pm, projectID, data)

	case "CellMove":
		return handleCellMove(pm, projectID, data)

	case "CellCreate":
		return handleCellCreate(pm, projectID, data)

	case "CellDelete":
		return handleCellDelete(pm, projectID, data)

	default:
		log.Printf("[Handler] Unknown packet type: %s", basePacket.Type)
		return nil, false
	}
}

// handleProjectSync processes ProjectSync packets
func handleProjectSync(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	// Parse the raw packet structure
	var rawPacket struct {
		BasePacket
		Project struct {
			ID        string            `json:"id"`
			Name      string            `json:"name"`
			Cells     []json.RawMessage `json:"cells"`
			CreatedAt time.Time         `json:"createdAt"`
			UpdatedAt time.Time         `json:"updatedAt"`
		} `json:"project"`
	}

	if err := json.Unmarshal(data, &rawPacket); err != nil {
		log.Printf("[Handler] Failed to parse ProjectSync packet: %v", err)
		return nil, false
	}

	// Parse cells with type detection
	cells, err := parseCells(rawPacket.Project.Cells)
	if err != nil {
		log.Printf("[Handler] Failed to parse cells: %v", err)
		return nil, false
	}

	// Create the project
	project := Project{
		ID:        rawPacket.Project.ID,
		Name:      rawPacket.Project.Name,
		Cells:     cells,
		CreatedAt: rawPacket.Project.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Update the project in the manager
	pm.SetProject(&project)

	log.Printf("[Handler] Project synced: %s (name: %s, cells: %d)", project.ID, project.Name, len(project.Cells))

	// Broadcast the updated project to all clients
	return data, true
}

// parseCells parses an array of raw JSON cells into concrete cell types
func parseCells(rawCells []json.RawMessage) ([]Cell, error) {
	cells := make([]Cell, 0, len(rawCells))

	for _, rawCell := range rawCells {
		// Determine cell type
		var cellType struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawCell, &cellType); err != nil {
			return nil, err
		}

		// Parse the appropriate cell type
		var cell Cell
		switch cellType.Type {
		case "equation":
			var eqCell EquationCell
			if err := json.Unmarshal(rawCell, &eqCell); err != nil {
				return nil, err
			}
			cell = eqCell

		case "folder":
			// Parse folder cell with nested cells
			var rawFolder struct {
				BaseCell
				Name  string            `json:"name"`
				Cells []json.RawMessage `json:"cells"`
			}
			if err := json.Unmarshal(rawCell, &rawFolder); err != nil {
				return nil, err
			}

			// Recursively parse nested cells
			nestedCells, err := parseCells(rawFolder.Cells)
			if err != nil {
				return nil, err
			}

			cell = FolderCell{
				BaseCell: rawFolder.BaseCell,
				Name:     rawFolder.Name,
				Cells:    nestedCells,
			}

		case "note":
			var noteCell NoteCell
			if err := json.Unmarshal(rawCell, &noteCell); err != nil {
				return nil, err
			}
			cell = noteCell

		default:
			log.Printf("[Handler] Unknown cell type: %s", cellType.Type)
			continue
		}

		cells = append(cells, cell)
	}

	return cells, nil
}

// handleCellUpdate processes CellUpdate packets
func handleCellUpdate(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	// Parse the raw packet to extract cell data
	var rawPacket struct {
		BasePacket
		CellID string          `json:"cellId"`
		Cell   json.RawMessage `json:"cell"`
	}

	if err := json.Unmarshal(data, &rawPacket); err != nil {
		log.Printf("[Handler] Failed to parse CellUpdate packet: %v", err)
		return nil, false
	}

	// Determine cell type
	var cellType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rawPacket.Cell, &cellType); err != nil {
		log.Printf("[Handler] Failed to parse cell type: %v", err)
		return nil, false
	}

	// Parse the appropriate cell type
	var cell Cell
	switch cellType.Type {
	case "equation":
		var eqCell EquationCell
		if err := json.Unmarshal(rawPacket.Cell, &eqCell); err != nil {
			log.Printf("[Handler] Failed to parse EquationCell: %v", err)
			return nil, false
		}
		cell = eqCell

	case "folder":
		var folderCell FolderCell
		if err := json.Unmarshal(rawPacket.Cell, &folderCell); err != nil {
			log.Printf("[Handler] Failed to parse FolderCell: %v", err)
			return nil, false
		}
		cell = folderCell

	case "note":
		var noteCell NoteCell
		if err := json.Unmarshal(rawPacket.Cell, &noteCell); err != nil {
			log.Printf("[Handler] Failed to parse NoteCell: %v", err)
			return nil, false
		}
		cell = noteCell

	default:
		log.Printf("[Handler] Unknown cell type: %s", cellType.Type)
		return nil, false
	}

	// Update the cell in the project
	if !pm.UpdateCell(projectID, rawPacket.CellID, cell) {
		log.Printf("[Handler] Failed to update cell: %s in project: %s", rawPacket.CellID, projectID)
		return nil, false
	}

	log.Printf("[Handler] Cell updated: %s (type: %s) in project: %s", rawPacket.CellID, cellType.Type, projectID)

	// Broadcast the update to all clients
	return data, true
}

// handleCellMove processes CellMove packets
func handleCellMove(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	var packet CellMovePacket
	if err := json.Unmarshal(data, &packet); err != nil {
		log.Printf("[Handler] Failed to parse CellMove packet: %v", err)
		return nil, false
	}

	// Move the cell in the project
	if !pm.MoveCell(projectID, packet.CellID, packet.FromIndex, packet.ToIndex) {
		log.Printf("[Handler] Failed to move cell: %s in project: %s", packet.CellID, projectID)
		return nil, false
	}

	log.Printf("[Handler] Cell moved: %s from %d to %d in project: %s", packet.CellID, packet.FromIndex, packet.ToIndex, projectID)

	// Broadcast the move to all clients
	return data, true
}

// handleCellCreate processes CellCreate packets
func handleCellCreate(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	// Parse the raw packet to extract cell data
	var rawPacket struct {
		BasePacket
		Cell         json.RawMessage `json:"cell"`
		ParentCellID *string         `json:"parentCellId,omitempty"`
		Index        int             `json:"index"`
	}

	if err := json.Unmarshal(data, &rawPacket); err != nil {
		log.Printf("[Handler] Failed to parse CellCreate packet: %v", err)
		return nil, false
	}

	// Determine cell type and parse
	var cellType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rawPacket.Cell, &cellType); err != nil {
		log.Printf("[Handler] Failed to parse cell type: %v", err)
		return nil, false
	}

	// Parse the appropriate cell type
	var cell Cell
	switch cellType.Type {
	case "equation":
		var eqCell EquationCell
		if err := json.Unmarshal(rawPacket.Cell, &eqCell); err != nil {
			log.Printf("[Handler] Failed to parse EquationCell: %v", err)
			return nil, false
		}
		cell = eqCell

	case "folder":
		// Parse folder cell with nested cells
		var rawFolder struct {
			BaseCell
			Name  string            `json:"name"`
			Cells []json.RawMessage `json:"cells"`
		}
		if err := json.Unmarshal(rawPacket.Cell, &rawFolder); err != nil {
			log.Printf("[Handler] Failed to parse FolderCell: %v", err)
			return nil, false
		}

		// Recursively parse nested cells
		nestedCells, err := parseCells(rawFolder.Cells)
		if err != nil {
			log.Printf("[Handler] Failed to parse nested cells: %v", err)
			return nil, false
		}

		cell = FolderCell{
			BaseCell: rawFolder.BaseCell,
			Name:     rawFolder.Name,
			Cells:    nestedCells,
		}

	case "note":
		var noteCell NoteCell
		if err := json.Unmarshal(rawPacket.Cell, &noteCell); err != nil {
			log.Printf("[Handler] Failed to parse NoteCell: %v", err)
			return nil, false
		}
		cell = noteCell

	default:
		log.Printf("[Handler] Unknown cell type: %s", cellType.Type)
		return nil, false
	}

	// Create the cell in the project
	if !pm.CreateCell(projectID, cell, rawPacket.ParentCellID, rawPacket.Index) {
		log.Printf("[Handler] Failed to create cell in project: %s", projectID)
		return nil, false
	}

	log.Printf("[Handler] Cell created: %s (type: %s) at index %d in project: %s", cell.GetID(), cellType.Type, rawPacket.Index, projectID)

	// Broadcast the creation to all clients
	return data, true
}

// handleCellDelete processes CellDelete packets
func handleCellDelete(pm *ProjectManager, projectID string, data []byte) ([]byte, bool) {
	var packet CellDeletePacket
	if err := json.Unmarshal(data, &packet); err != nil {
		log.Printf("[Handler] Failed to parse CellDelete packet: %v", err)
		return nil, false
	}

	// Delete the cell from the project
	if !pm.DeleteCell(projectID, packet.CellID, packet.ParentCellID) {
		log.Printf("[Handler] Failed to delete cell: %s in project: %s", packet.CellID, projectID)
		return nil, false
	}

	log.Printf("[Handler] Cell deleted: %s in project: %s", packet.CellID, projectID)

	// Broadcast the deletion to all clients
	return data, true
}
