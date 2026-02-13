package workflow

import (
	"fmt"

	"github.com/Wangren-Academy/Agent/backend/internal/store"

	"github.com/google/uuid"
)

// Node represents a node in the workflow DAG
type Node struct {
	ID         string
	AgentID    uuid.UUID
	AgentName  string
	Position   store.Position
	InputMap   map[string]string
	Config     map[string]any
	DependsOn  []string
	Downstream []string
}

// Edge represents a connection between nodes
type Edge struct {
	ID     string
	Source string
	Target string
}

// DAG represents the workflow as a Directed Acyclic Graph
type DAG struct {
	Nodes     map[string]*Node
	Edges     []*Edge
	InDegrees map[string]int
	OutEdges  map[string][]string
}

// NewDAG creates a new DAG from workflow configuration
func NewDAG(workflow *store.Workflow) (*DAG, error) {
	dag := &DAG{
		Nodes:     make(map[string]*Node),
		Edges:     make([]*Edge, 0),
		InDegrees: make(map[string]int),
		OutEdges:  make(map[string][]string),
	}

	// Add nodes
	for _, nodeConfig := range workflow.Nodes {
		node := &Node{
			ID:         nodeConfig.ID,
			AgentID:    nodeConfig.AgentID,
			Position:   nodeConfig.Position,
			Config:     nodeConfig.Data,
			DependsOn:  make([]string, 0),
			Downstream: make([]string, 0),
		}
		dag.Nodes[node.ID] = node
		dag.InDegrees[node.ID] = 0
		dag.OutEdges[node.ID] = make([]string, 0)
	}

	// Add edges
	for _, edgeConfig := range workflow.Edges {
		edge := &Edge{
			ID:     edgeConfig.ID,
			Source: edgeConfig.Source,
			Target: edgeConfig.Target,
		}
		dag.Edges = append(dag.Edges, edge)

		// Build adjacency lists
		dag.OutEdges[edge.Source] = append(dag.OutEdges[edge.Source], edge.Target)
		dag.Nodes[edge.Target].DependsOn = append(dag.Nodes[edge.Target].DependsOn, edge.Source)
		dag.InDegrees[edge.Target]++
		dag.Nodes[edge.Source].Downstream = append(dag.Nodes[edge.Source].Downstream, edge.Target)
	}

	// Validate DAG (no cycles)
	if err := dag.Validate(); err != nil {
		return nil, err
	}

	return dag, nil
}

// Validate checks if the DAG is valid (no cycles)
func (d *DAG) Validate() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range d.Nodes {
		if !visited[nodeID] {
			if d.hasCycle(nodeID, visited, recStack) {
				return fmt.Errorf("cycle detected in workflow DAG")
			}
		}
	}
	return nil
}

func (d *DAG) hasCycle(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range d.OutEdges[nodeID] {
		if !visited[neighbor] {
			if d.hasCycle(neighbor, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// TopologicalSort returns nodes in topological order
func (d *DAG) TopologicalSort() []string {
	inDegree := make(map[string]int)
	for k, v := range d.InDegrees {
		inDegree[k] = v
	}

	queue := make([]string, 0)
	result := make([]string, 0)

	// Find all nodes with no incoming edges
	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		for _, neighbor := range d.OutEdges[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return result
}

// GetReadyNodes returns nodes that are ready to execute
func (d *DAG) GetReadyNodes(completed map[string]bool) []string {
	ready := make([]string, 0)
	for nodeID, node := range d.Nodes {
		if completed[nodeID] {
			continue
		}
		allDepsComplete := true
		for _, dep := range node.DependsOn {
			if !completed[dep] {
				allDepsComplete = false
				break
			}
		}
		if allDepsComplete {
			ready = append(ready, nodeID)
		}
	}
	return ready
}
