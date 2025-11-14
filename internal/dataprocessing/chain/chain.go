// Package chain 数据处理责任链模式
// 用于数据清洗、特征提取等任务的链式处理
package chain

import (
	"context"
	"fmt"
)

// DataProcessor 数据处理接口
type DataProcessor interface {
	Process(ctx context.Context, data interface{}) (interface{}, error)
	CanProcess(data interface{}) bool
	GetName() string
}

// ProcessingChain 责任链
type ProcessingChain struct {
	processors []DataProcessor
}

// NewProcessingChain 创建新的处理链
func NewProcessingChain(processors ...DataProcessor) *ProcessingChain {
	return &ProcessingChain{
		processors: processors,
	}
}

// Process 执行处理链
func (c *ProcessingChain) Process(ctx context.Context, data interface{}) (interface{}, error) {
	var err error
	result := data
	
	for _, processor := range c.processors {
		if processor.CanProcess(result) {
			result, err = processor.Process(ctx, result)
			if err != nil {
				return nil, fmt.Errorf("处理器 %s 失败: %w", processor.GetName(), err)
			}
		}
	}
	
	return result, nil
}

// AddProcessor 添加处理器到链中
func (c *ProcessingChain) AddProcessor(processor DataProcessor) {
	c.processors = append(c.processors, processor)
}

// GetProcessors 获取所有处理器
func (c *ProcessingChain) GetProcessors() []DataProcessor {
	return c.processors
}