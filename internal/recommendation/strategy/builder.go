// Package strategy 推荐排序策略构建器
package strategy

// StrategyBuilder 策略构建器
type StrategyBuilder struct {
	strategies []RankingStrategy
}

// NewStrategyBuilder 创建策略构建器
func NewStrategyBuilder() *StrategyBuilder {
	return &StrategyBuilder{
		strategies: make([]RankingStrategy, 0),
	}
}

// WithScoreBased 添加基于分数的策略
func (b *StrategyBuilder) WithScoreBased() *StrategyBuilder {
	b.strategies = append(b.strategies, NewScoreBasedStrategy())
	return b
}

// WithDiversity 添加多样性策略
func (b *StrategyBuilder) WithDiversity() *StrategyBuilder {
	b.strategies = append(b.strategies, NewDiversityStrategy())
	return b
}

// WithNovelty 添加新颖性策略
func (b *StrategyBuilder) WithNovelty() *StrategyBuilder {
	b.strategies = append(b.strategies, NewNoveltyStrategy())
	return b
}

// WithPersonalization 添加个性化策略
func (b *StrategyBuilder) WithPersonalization() *StrategyBuilder {
	b.strategies = append(b.strategies, NewPersonalizationStrategy())
	return b
}

// Build 构建策略组合
func (b *StrategyBuilder) Build() []RankingStrategy {
	return b.strategies
}

// BuildDefaultStrategies 构建默认策略组合
func BuildDefaultStrategies() []RankingStrategy {
	return NewStrategyBuilder().
		WithScoreBased().
		WithDiversity().
		Build()
}

// BuildEcommerceStrategies 构建电商推荐策略组合
func BuildEcommerceStrategies() []RankingStrategy {
	return NewStrategyBuilder().
		WithScoreBased().
		WithDiversity().
		WithNovelty().
		Build()
}

// BuildContentStrategies 构建内容推荐策略组合
func BuildContentStrategies() []RankingStrategy {
	return NewStrategyBuilder().
		WithScoreBased().
		WithPersonalization().
		Build()
}