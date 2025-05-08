package snowflake

import (
	"sync"
	"testing"
	"time"
)

func TestSnowflakeIDGenerator_NextID(t *testing.T) {
	// 创建ID生成器
	generator, err := NewSnowflakeIDGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create ID generator: %v", err)
	}

	// 测试生成ID的唯一性
	t.Run("Uniqueness", func(t *testing.T) {
		count := 10000
		ids := make(map[int64]bool)

		for i := 0; i < count; i++ {
			id, err := generator.NextID("user1")
			if err != nil {
				t.Fatalf("Failed to generate ID: %v", err)
			}

			if _, exists := ids[id]; exists {
				t.Fatalf("Duplicate ID generated: %d", id)
			}

			ids[id] = true
		}
	})

	// 测试不同用户ID生成的ID不同
	t.Run("Different UserIDs", func(t *testing.T) {
		id1, err := generator.NextID("user1")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		id2, err := generator.NextID("user2")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if id1 == id2 {
			t.Fatalf("IDs for different users should be different, got %d for both", id1)
		}
	})

	// 测试时间排序性
	t.Run("Time Ordering", func(t *testing.T) {
		id1, err := generator.NextID("user1")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		time.Sleep(5 * time.Millisecond)

		id2, err := generator.NextID("user1")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if id2 <= id1 {
			t.Fatalf("Later ID should be greater than earlier ID, got %d <= %d", id2, id1)
		}
	})

	// 测试并发安全性
	t.Run("Concurrency Safety", func(t *testing.T) {
		count := 1000
		ids := sync.Map{}
		wg := sync.WaitGroup{}
		wg.Add(count)

		for i := 0; i < count; i++ {
			go func(i int) {
				defer wg.Done()
				id, err := generator.NextID("user1")
				if err != nil {
					t.Errorf("Failed to generate ID: %v", err)
					return
				}

				if _, loaded := ids.LoadOrStore(id, true); loaded {
					t.Errorf("Duplicate ID generated: %d", id)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestInvalidNodeID(t *testing.T) {
	// 测试节点ID超出范围
	_, err := NewSnowflakeIDGenerator(-1)
	if err == nil {
		t.Fatalf("Expected error for negative node ID, but got nil")
	}

	_, err = NewSnowflakeIDGenerator(1024)
	if err == nil {
		t.Fatalf("Expected error for excessive node ID, but got nil")
	}
}
