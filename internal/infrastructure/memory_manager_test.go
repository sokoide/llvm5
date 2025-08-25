package infrastructure

import (
	"testing"
)

// TestPooledMemoryManager tests pooled memory manager basic functionality
func TestPooledMemoryManager(t *testing.T) {
	manager := NewPooledMemoryManager()
	if manager == nil {
		t.Fatal("NewPooledMemoryManager should return non-nil manager")
	}

	// Test initial stats
	stats := manager.GetStats()
	if stats.TotalMemoryUsed < 0 {
		t.Error("Initial TotalMemoryUsed should be non-negative")
	}

	// Test that methods don't panic
	node, err := manager.AllocateNode("test", 64)
	if err != nil {
		t.Errorf("AllocateNode should not fail: %v", err)
	}
	if node == nil {
		t.Error("AllocateNode should return non-nil pointer")
	}

	str, err := manager.AllocateString("test")
	if err != nil {
		t.Errorf("AllocateString should not fail: %v", err)
	}
	if str == nil {
		t.Error("AllocateString should return non-nil")
	}

	// Test cleanup
	manager.FreeAll()

	finalStats := manager.GetStats()
	if finalStats.TotalMemoryUsed > stats.TotalMemoryUsed {
		t.Log("Memory usage after FreeAll:", finalStats.TotalMemoryUsed)
	}
}

// TestCompactMemoryManager tests compact memory manager
func TestCompactMemoryManager(t *testing.T) {
	manager := NewCompactMemoryManager()
	if manager == nil {
		t.Fatal("NewCompactMemoryManager should return non-nil manager")
	}

	// Test basic operations
	node, err := manager.AllocateNode("test", 128)
	if err != nil {
		t.Errorf("AllocateNode should not fail: %v", err)
	}
	if node == nil {
		t.Error("AllocateNode should return non-nil pointer")
	}

	str, err := manager.AllocateString("compact test")
	if err != nil {
		t.Errorf("AllocateString should not fail: %v", err)
	}
	if str == nil {
		t.Error("AllocateString should return non-nil")
	}

	// Test cleanup
	manager.FreeAll()
}

// TestTrackingMemoryManager tests tracking memory manager
func TestTrackingMemoryManager(t *testing.T) {
	underlying := NewPooledMemoryManager()
	manager := NewTrackingMemoryManager(underlying)
	if manager == nil {
		t.Fatal("NewTrackingMemoryManager should return non-nil manager")
	}

	// Test basic operations
	node, err := manager.AllocateNode("tracked", 256)
	if err != nil {
		t.Errorf("AllocateNode should not fail: %v", err)
	}
	if node == nil {
		t.Error("AllocateNode should return non-nil pointer")
	}

	str, err := manager.AllocateString("tracked string")
	if err != nil {
		t.Errorf("AllocateString should not fail: %v", err)
	}
	if str == nil {
		t.Error("AllocateString should return non-nil")
	}

	// Test allocation log
	allocationLog := manager.GetAllocationLog()
	if len(allocationLog) == 0 {
		t.Error("Allocation log should contain entries after allocations")
	}

	// Test memory report (should not panic)
	manager.PrintMemoryReport()

	// Test cleanup
	manager.FreeAll()
}

// TestMemoryManagerStats tests memory statistics functionality
func TestMemoryManagerStats(t *testing.T) {
	pooled := NewPooledMemoryManager()
	compact := NewCompactMemoryManager()
	tracking := NewTrackingMemoryManager(NewPooledMemoryManager())

	managers := []interface{}{
		pooled,
		compact,
		tracking,
	}

	managerNames := []string{"Pooled", "Compact", "Tracking"}

	for i, mgr := range managers {
		t.Run(managerNames[i], func(t *testing.T) {
			// Test basic functionality based on type
			switch v := mgr.(type) {
			case *PooledMemoryManager:
				// Test PooledMemoryManager
				stats := v.GetStats()
				if stats.TotalMemoryUsed < 0 {
					t.Error("Initial TotalMemoryUsed should be non-negative")
				}
				v.FreeAll()

			case *CompactMemoryManager:
				// Test CompactMemoryManager
				stats := v.GetStats()
				if stats.TotalMemoryUsed < 0 {
					t.Error("Initial TotalMemoryUsed should be non-negative")
				}
				v.FreeAll()

			case *TrackingMemoryManager:
				// Test TrackingMemoryManager
				stats := v.GetStats()
				if stats.TotalMemoryUsed < 0 {
					t.Error("Initial TotalMemoryUsed should be non-negative")
				}
				v.FreeAll()

			default:
				t.Skip("Unknown manager type")
			}
		})
	}
}
