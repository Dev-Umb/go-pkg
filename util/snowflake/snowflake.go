// Package snowflake 实现基于雪花算法的唯一ID生成器
/**
* @Author: Game Room Team
* @Date: 2023-06-15
 */
package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	// 起始时间戳（2023-01-01 00:00:00 UTC）
	twepoch = int64(1672531200000)

	// 各部分的位数
	nodeIDBits   = 10 // 节点ID占用的位数
	sequenceBits = 12 // 序列号占用的位数

	// 各部分的最大值
	maxNodeID   = -1 ^ (-1 << nodeIDBits)   // 最大节点ID
	maxSequence = -1 ^ (-1 << sequenceBits) // 最大序列号

	// 各部分的左移位数
	nodeIDShift    = sequenceBits              // 节点ID左移位数
	timestampShift = sequenceBits + nodeIDBits // 时间戳左移位数
)

// Generator 定义雪花算法ID生成器接口
type Generator interface {
	NextID(userID string) (int64, error)
}

// SnowflakeIDGenerator 雪花算法ID生成器实现
type SnowflakeIDGenerator struct {
	mu       sync.Mutex
	nodeID   int64 // 节点ID
	sequence int64 // 序列号
	lastTime int64 // 上次生成ID的时间戳
}

// NewSnowflakeIDGenerator 创建一个新的雪花算法ID生成器
func NewSnowflakeIDGenerator(nodeID int64) (*SnowflakeIDGenerator, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, errors.New("node ID must be between 0 and 1023")
	}
	return &SnowflakeIDGenerator{
		nodeID:   nodeID,
		sequence: 0,
		lastTime: 0,
	}, nil
}

// NextID 生成下一个唯一ID
func (s *SnowflakeIDGenerator) NextID(userID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取当前时间戳
	now := time.Now().UnixNano() / 1e6 // 转换为毫秒

	// 如果当前时间小于上次生成ID的时间，说明系统时钟回退，拒绝生成ID
	if now < s.lastTime {
		return 0, errors.New("clock moved backwards, refusing to generate ID")
	}

	// 如果是同一时间生成的，则进行序列号递增
	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & maxSequence
		// 如果序列号溢出，则等待下一毫秒
		if s.sequence == 0 {
			// 阻塞到下一个毫秒
			for now <= s.lastTime {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		// 如果是新的时间戳，则重置序列号
		s.sequence = 0
	}

	// 更新lastTime
	s.lastTime = now

	// 用户ID的哈希值，作为额外的随机性来源
	userHash := int64(0)
	if userID != "" {
		for _, ch := range userID {
			userHash = (userHash*31 + int64(ch)) & maxNodeID
		}
	}

	// 节点ID与用户哈希异或，提供更好的分布性
	nodeWithUserHash := (s.nodeID ^ userHash) & maxNodeID

	// 生成ID
	id := ((now - twepoch) << timestampShift) |
		(nodeWithUserHash << nodeIDShift) |
		s.sequence

	return id, nil
}
