// Package acl 防腐层（Anti-Corruption Layer）
// 用于隔离外部系统变化对核心业务的影响
package acl

import (
	"context"
	"time"
)