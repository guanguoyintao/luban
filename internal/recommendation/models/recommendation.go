package models

import (
	"time"
)

// 用户模型
type User struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Demographics map[string]interface{} `json:"demographics"`
	Preferences  map[string]interface{} `json:"preferences"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 物品模型
type Item struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Features    map[string]interface{} `json:"features"`
	Price       float64                `json:"price"`
	Rating      float64                `json:"rating"`
	Popularity  int                    `json:"popularity"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 用户行为模型
type UserBehavior struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	ItemID    string                 `json:"item_id"`
	Type      string                 `json:"type"` // click, view, purchase, rating, etc.
	Value     float64                `json:"value"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// 推荐结果模型
type Recommendation struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ItemID       string                 `json:"item_id"`
	Score        float64                `json:"score"`
	Confidence   float64                `json:"confidence"`
	Algorithm    string                 `json:"algorithm"`
	Reason       string                 `json:"reason"`
	Context      map[string]interface{} `json:"context"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// 推荐会话模型
type RecommendationSession struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Scenario    string                 `json:"scenario"`
	RequestID   string                 `json:"request_id"`
	Context     map[string]interface{} `json:"context"`
	Status      string                 `json:"status"` // active, completed, expired
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 用户反馈模型
type UserFeedback struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	RecommendationID string                `json:"recommendation_id"`
	ItemID          string                 `json:"item_id"`
	Type            string                 `json:"type"` // like, dislike, click, ignore, etc.
	Value           float64                `json:"value"`
	Context         map[string]interface{} `json:"context"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// 算法性能指标模型
type AlgorithmMetrics struct {
	Algorithm       string                 `json:"algorithm"`
	Precision       float64                `json:"precision"`
	Recall          float64                `json:"recall"`
	F1Score         float64                `json:"f1_score"`
	CTR             float64                `json:"ctr"`              // Click-through rate
	ConversionRate  float64                `json:"conversion_rate"`
	AverageScore    float64                `json:"average_score"`
	Coverage        float64                `json:"coverage"`
	Diversity       float64                `json:"diversity"`
	Novelty         float64                `json:"novelty"`
	TimePeriod      string                 `json:"time_period"`
	CalculatedAt    time.Time              `json:"calculated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// A/B测试模型
type ABTest struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Algorithms      []string               `json:"algorithms"`
	TrafficSplit    map[string]float64     `json:"traffic_split"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Status          string                 `json:"status"` // active, completed, paused
	Metrics         map[string]interface{} `json:"metrics"`
	Winner          string                 `json:"winner"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// 特征工程模型
type Feature struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // numerical, categorical, text, etc.
	Value       interface{}            `json:"value"`
	Importance  float64                `json:"importance"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 模型训练结果模型
type ModelTrainingResult struct {
	ID              string                 `json:"id"`
	Algorithm       string                 `json:"algorithm"`
	ModelVersion    string                 `json:"model_version"`
	TrainingData    string                 `json:"training_data"`
	ValidationData  string                 `json:"validation_data"`
	TestData        string                 `json:"test_data"`
	Metrics         map[string]interface{} `json:"metrics"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	TrainingTime    int64                  `json:"training_time"`
	Status          string                 `json:"status"` // success, failed, running
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time"`
	ModelPath       string                 `json:"model_path"`
	Metadata        map[string]interface{} `json:"metadata"`
}