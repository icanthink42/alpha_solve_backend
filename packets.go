package main

import (
	"encoding/json"
	"time"
)

// Packet is an interface that all packet types must implement
type Packet interface {
	GetType() string
	GetTimestamp() string
	isPacket() // private method to ensure only our types implement this
}

// BasePacket contains common fields for all packet types
type BasePacket struct {
	Type      string `json:"type"`      // 'ProjectSync' | 'CellUpdate' | 'CellMove' | 'CellCreate' | 'CellDelete'
	Timestamp string `json:"timestamp"` // ISO 8601 timestamp
}

// GetType returns the packet type
func (b BasePacket) GetType() string {
	return b.Type
}

// GetTimestamp returns the timestamp
func (b BasePacket) GetTimestamp() string {
	return b.Timestamp
}

// ProjectSyncPacket syncs the entire project
type ProjectSyncPacket struct {
	BasePacket
	Project Project `json:"project"`
}

func (p ProjectSyncPacket) isPacket() {}

// CellUpdatePacket updates a cell with a specific id
type CellUpdatePacket struct {
	BasePacket
	CellID string `json:"cellId"`
	Cell   Cell   `json:"cell"`
}

func (c CellUpdatePacket) isPacket() {}

// CellMovePacket moves a cell from one index to another
type CellMovePacket struct {
	BasePacket
	CellID    string `json:"cellId"`
	FromIndex int    `json:"fromIndex"`
	ToIndex   int    `json:"toIndex"`
}

func (c CellMovePacket) isPacket() {}

// CellCreatePacket creates a new cell at a specific position
type CellCreatePacket struct {
	BasePacket
	Cell         Cell    `json:"cell"`
	ParentCellID *string `json:"parentCellId,omitempty"` // Optional: if creating inside a folder
	Index        int     `json:"index"`                  // Position to insert the cell
}

func (c CellCreatePacket) isPacket() {}

// CellDeletePacket deletes a cell by ID
type CellDeletePacket struct {
	BasePacket
	CellID       string  `json:"cellId"`
	ParentCellID *string `json:"parentCellId,omitempty"` // Optional: if deleting from inside a folder
}

func (c CellDeletePacket) isPacket() {}

// WebSocketConfig represents connection configuration for websocket
type WebSocketConfig struct {
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
	UserID    string `json:"userId"`
	URL       string `json:"url,omitempty"` // Optional, defaults to ws://localhost:8080
}

// PacketFactory provides utility functions for packet creation

// CreateProjectSync creates a ProjectSync packet
func CreateProjectSync(project Project) ProjectSyncPacket {
	return ProjectSyncPacket{
		BasePacket: BasePacket{
			Type:      "ProjectSync",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Project: project,
	}
}

// CreateCellUpdate creates a CellUpdate packet
func CreateCellUpdate(cellID string, cell Cell) CellUpdatePacket {
	return CellUpdatePacket{
		BasePacket: BasePacket{
			Type:      "CellUpdate",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		CellID: cellID,
		Cell:   cell,
	}
}

// CreateCellMove creates a CellMove packet
func CreateCellMove(cellID string, fromIndex, toIndex int) CellMovePacket {
	return CellMovePacket{
		BasePacket: BasePacket{
			Type:      "CellMove",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		CellID:    cellID,
		FromIndex: fromIndex,
		ToIndex:   toIndex,
	}
}

// CreateCellCreate creates a CellCreate packet
func CreateCellCreate(cell Cell, parentCellID *string, index int) CellCreatePacket {
	return CellCreatePacket{
		BasePacket: BasePacket{
			Type:      "CellCreate",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Cell:         cell,
		ParentCellID: parentCellID,
		Index:        index,
	}
}

// CreateCellDelete creates a CellDelete packet
func CreateCellDelete(cellID string, parentCellID *string) CellDeletePacket {
	return CellDeletePacket{
		BasePacket: BasePacket{
			Type:      "CellDelete",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		CellID:       cellID,
		ParentCellID: parentCellID,
	}
}

// SerializePacket serializes a packet to JSON string
func SerializePacket(packet Packet) (string, error) {
	bytes, err := json.Marshal(packet)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DeserializePacket deserializes a JSON string to a packet
// Note: This returns a map because Go's type system requires knowing the concrete type at compile time
// You'll need to inspect the "type" field to determine which packet type to unmarshal into
func DeserializePacket(jsonStr string) (map[string]interface{}, error) {
	var packet map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &packet)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

// DeserializeProjectSyncPacket deserializes a JSON string into a ProjectSyncPacket
func DeserializeProjectSyncPacket(jsonStr string) (*ProjectSyncPacket, error) {
	var packet ProjectSyncPacket
	err := json.Unmarshal([]byte(jsonStr), &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}

// DeserializeCellUpdatePacket deserializes a JSON string into a CellUpdatePacket
func DeserializeCellUpdatePacket(jsonStr string) (*CellUpdatePacket, error) {
	var packet CellUpdatePacket
	err := json.Unmarshal([]byte(jsonStr), &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}

// DeserializeCellMovePacket deserializes a JSON string into a CellMovePacket
func DeserializeCellMovePacket(jsonStr string) (*CellMovePacket, error) {
	var packet CellMovePacket
	err := json.Unmarshal([]byte(jsonStr), &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}
