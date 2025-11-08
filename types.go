package main

import "time"

// Project represents a project containing cells
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Cells     []Cell    `json:"cells"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Context represents the context for an equation cell
type Context struct {
	// Add context fields as needed
}

// Dropdown represents a dropdown option
type Dropdown struct {
	// Add dropdown fields as needed
}

// DropdownSelection represents a selected dropdown value
type DropdownSelection struct {
	// Add dropdown selection fields as needed
}

// Cell is an interface that all cell types must implement
type Cell interface {
	GetID() string
	GetType() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	isCell() // private method to ensure only our types implement this
}

// BaseCell contains common fields for all cell types
type BaseCell struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // 'equation' | 'folder' | 'note'
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetID returns the cell ID
func (b BaseCell) GetID() string {
	return b.ID
}

// GetType returns the cell type
func (b BaseCell) GetType() string {
	return b.Type
}

// GetCreatedAt returns the creation timestamp
func (b BaseCell) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns the update timestamp
func (b BaseCell) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// EquationCell represents a cell containing a LaTeX equation
type EquationCell struct {
	BaseCell
	Latex              string              `json:"latex"`
	Context            *Context            `json:"context,omitempty"`
	Solutions          []string            `json:"solutions,omitempty"`
	Dropdowns          []Dropdown          `json:"dropdowns,omitempty"`
	DropdownSelections []DropdownSelection `json:"dropdownSelections,omitempty"`
}

func (e EquationCell) isCell() {}

// FolderCell represents a cell that acts as a folder containing other cells
type FolderCell struct {
	BaseCell
	Name  string `json:"name"`
	Cells []Cell `json:"cells"`
}

func (f FolderCell) isCell() {}

// NoteCell represents a cell containing plain text notes
type NoteCell struct {
	BaseCell
	Content string `json:"content"`
}

func (n NoteCell) isCell() {}
