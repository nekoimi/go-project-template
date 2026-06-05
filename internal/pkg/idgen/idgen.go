package idgen

import (
	"errors"
	"fmt"
	"strconv"

	sf "github.com/bwmarrin/snowflake"
)

var (
	node        *sf.Node
	ErrInvalidID = errors.New("invalid id")
)

// Init 初始化雪花算法节点
func Init(nodeID int64) error {
	var err error
	node, err = sf.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to create snowflake node: %w", err)
	}
	return nil
}

// GenerateID 生成雪花 ID
func GenerateID() int64 {
	return node.Generate().Int64()
}

// GenerateStringID 生成雪花 ID 字符串
func GenerateStringID() string {
	return node.Generate().String()
}

// FormatID converts numeric snowflake ID to string.
func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// ParseID parses decimal string snowflake ID into int64.
func ParseID(id string) (int64, error) {
	if id == "" {
		return 0, ErrInvalidID
	}

	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, ErrInvalidID
	}

	return parsed, nil
}
