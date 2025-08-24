// Package infrastructure contains memory management implementation
package infrastructure

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/sokoide/llvm5/staticlang/internal/interfaces"
)

// PooledMemoryManager implements MemoryManager using memory pools for efficiency
type PooledMemoryManager struct {
	mutex            sync.RWMutex
	nodePools        map[string]*nodePool
	stringPool       map[string]*stringEntry
	totalAllocated   int
	nodesAllocated   int
	stringsAllocated int
}

type nodePool struct {
	nodes []interface{}
	size  int
}

type stringEntry struct {
	value    string
	refCount int
}

// NewPooledMemoryManager creates a new pooled memory manager
func NewPooledMemoryManager() *PooledMemoryManager {
	return &PooledMemoryManager{
		nodePools:  make(map[string]*nodePool),
		stringPool: make(map[string]*stringEntry),
	}
}

// AllocateNode allocates memory for an AST node using type-specific pools
func (mm *PooledMemoryManager) AllocateNode(nodeType string, size int) (interface{}, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	pool, exists := mm.nodePools[nodeType]
	if !exists {
		pool = &nodePool{
			nodes: make([]interface{}, 0, 16), // Start with small capacity
			size:  size,
		}
		mm.nodePools[nodeType] = pool
	}

	// For simplicity, we'll allocate a new node each time
	// In a real implementation, you might want to reuse nodes from the pool
	var node interface{}

	switch nodeType {
	case "LiteralExpr":
		node = &struct{}{} // Placeholder - would be actual node type
	case "BinaryExpr":
		node = &struct{}{} // Placeholder - would be actual node type
	case "IdentifierExpr":
		node = &struct{}{} // Placeholder - would be actual node type
	default:
		node = make([]byte, size) // Generic allocation
	}

	pool.nodes = append(pool.nodes, node)
	mm.nodesAllocated++
	mm.totalAllocated += size

	return node, nil
}

// AllocateString allocates memory for a string with reference counting
func (mm *PooledMemoryManager) AllocateString(s string) (interface{}, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if entry, exists := mm.stringPool[s]; exists {
		entry.refCount++
		return entry.value, nil
	}

	entry := &stringEntry{
		value:    s,
		refCount: 1,
	}

	mm.stringPool[s] = entry
	mm.stringsAllocated++
	mm.totalAllocated += len(s)

	return entry.value, nil
}

// ReleaseString decrements the reference count for a string
func (mm *PooledMemoryManager) ReleaseString(s string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if entry, exists := mm.stringPool[s]; exists {
		entry.refCount--
		if entry.refCount <= 0 {
			mm.totalAllocated -= len(s)
			delete(mm.stringPool, s)
		}
	}
}

// FreeAll frees all allocated memory
func (mm *PooledMemoryManager) FreeAll() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Clear node pools
	for nodeType := range mm.nodePools {
		delete(mm.nodePools, nodeType)
	}

	// Clear string pool
	for s := range mm.stringPool {
		delete(mm.stringPool, s)
	}

	mm.totalAllocated = 0
	mm.nodesAllocated = 0
	mm.stringsAllocated = 0
}

// GetStats returns memory usage statistics
func (mm *PooledMemoryManager) GetStats() interfaces.MemoryStats {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return interfaces.MemoryStats{
		NodesAllocated:   mm.nodesAllocated,
		StringsAllocated: mm.stringsAllocated,
		TotalMemoryUsed:  mm.totalAllocated,
	}
}

// GetPoolStats returns detailed statistics about memory pools
func (mm *PooledMemoryManager) GetPoolStats() map[string]int {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	stats := make(map[string]int)
	for nodeType, pool := range mm.nodePools {
		stats[nodeType] = len(pool.nodes)
	}
	return stats
}

// CompactMemoryManager is a simpler memory manager that doesn't use pools
type CompactMemoryManager struct {
	mutex       sync.RWMutex
	allocations []allocation
	totalMemory int
}

type allocation struct {
	ptr        interface{}
	size       int
	objectType string
}

// NewCompactMemoryManager creates a new compact memory manager
func NewCompactMemoryManager() *CompactMemoryManager {
	return &CompactMemoryManager{
		allocations: make([]allocation, 0),
	}
}

// AllocateNode allocates memory for an AST node
func (mm *CompactMemoryManager) AllocateNode(nodeType string, size int) (interface{}, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Allocate the requested memory
	ptr := make([]byte, size)

	allocation := allocation{
		ptr:        ptr,
		size:       size,
		objectType: nodeType,
	}

	mm.allocations = append(mm.allocations, allocation)
	mm.totalMemory += size

	return ptr, nil
}

// AllocateString allocates memory for a string
func (mm *CompactMemoryManager) AllocateString(s string) (interface{}, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	size := len(s)
	allocation := allocation{
		ptr:        s,
		size:       size,
		objectType: "string",
	}

	mm.allocations = append(mm.allocations, allocation)
	mm.totalMemory += size

	return s, nil
}

// FreeAll frees all allocated memory
func (mm *CompactMemoryManager) FreeAll() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.allocations = mm.allocations[:0]
	mm.totalMemory = 0
}

// GetStats returns memory usage statistics
func (mm *CompactMemoryManager) GetStats() interfaces.MemoryStats {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	nodeCount := 0
	stringCount := 0

	for _, alloc := range mm.allocations {
		switch alloc.objectType {
		case "string":
			stringCount++
		default:
			nodeCount++
		}
	}

	return interfaces.MemoryStats{
		NodesAllocated:   nodeCount,
		StringsAllocated: stringCount,
		TotalMemoryUsed:  mm.totalMemory,
	}
}

// TrackingMemoryManager wraps another MemoryManager and provides detailed tracking
type TrackingMemoryManager struct {
	underlying    interfaces.MemoryManager
	allocationLog []AllocationEvent
	mutex         sync.RWMutex
}

type AllocationEvent struct {
	Type      string
	Size      int
	Timestamp int64
	Action    string // "allocate" or "free"
}

// NewTrackingMemoryManager creates a new tracking memory manager
func NewTrackingMemoryManager(underlying interfaces.MemoryManager) *TrackingMemoryManager {
	return &TrackingMemoryManager{
		underlying:    underlying,
		allocationLog: make([]AllocationEvent, 0),
	}
}

// AllocateNode allocates memory for an AST node and logs the allocation
func (mm *TrackingMemoryManager) AllocateNode(nodeType string, size int) (interface{}, error) {
	result, err := mm.underlying.AllocateNode(nodeType, size)
	if err != nil {
		return nil, err
	}

	mm.mutex.Lock()
	mm.allocationLog = append(mm.allocationLog, AllocationEvent{
		Type:   nodeType,
		Size:   size,
		Action: "allocate",
	})
	mm.mutex.Unlock()

	return result, nil
}

// AllocateString allocates memory for a string and logs the allocation
func (mm *TrackingMemoryManager) AllocateString(s string) (interface{}, error) {
	result, err := mm.underlying.AllocateString(s)
	if err != nil {
		return nil, err
	}

	mm.mutex.Lock()
	mm.allocationLog = append(mm.allocationLog, AllocationEvent{
		Type:   "string",
		Size:   len(s),
		Action: "allocate",
	})
	mm.mutex.Unlock()

	return result, nil
}

// FreeAll frees all allocated memory and logs the event
func (mm *TrackingMemoryManager) FreeAll() {
	mm.underlying.FreeAll()

	mm.mutex.Lock()
	mm.allocationLog = append(mm.allocationLog, AllocationEvent{
		Type:   "all",
		Size:   0,
		Action: "free",
	})
	mm.mutex.Unlock()
}

// GetStats returns memory usage statistics
func (mm *TrackingMemoryManager) GetStats() interfaces.MemoryStats {
	return mm.underlying.GetStats()
}

// GetAllocationLog returns the allocation log
func (mm *TrackingMemoryManager) GetAllocationLog() []AllocationEvent {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	log := make([]AllocationEvent, len(mm.allocationLog))
	copy(log, mm.allocationLog)
	return log
}

// PrintMemoryReport prints a detailed memory usage report
func (mm *TrackingMemoryManager) PrintMemoryReport() {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	stats := mm.GetStats()
	fmt.Printf("Memory Usage Report:\n")
	fmt.Printf("  Nodes Allocated: %d\n", stats.NodesAllocated)
	fmt.Printf("  Strings Allocated: %d\n", stats.StringsAllocated)
	fmt.Printf("  Total Memory Used: %d bytes\n", stats.TotalMemoryUsed)
	fmt.Printf("  Total Allocation Events: %d\n", len(mm.allocationLog))

	// Count allocations by type
	typeCounts := make(map[string]int)
	for _, event := range mm.allocationLog {
		if event.Action == "allocate" {
			typeCounts[event.Type]++
		}
	}

	fmt.Printf("  Allocations by Type:\n")
	for nodeType, count := range typeCounts {
		fmt.Printf("    %s: %d\n", nodeType, count)
	}
}

// GetMemorySize returns the size in bytes of a given object
func GetMemorySize(obj interface{}) int {
	return int(unsafe.Sizeof(obj))
}
