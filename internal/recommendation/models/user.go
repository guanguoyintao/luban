package models

import (
	"time"
)

// 用户数据模型
type UserData struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Age         int                    `json:"age"`
	Gender      string                 `json:"gender"`
	Location    string                 `json:"location"`
	Occupation  string                 `json:"occupation"`
	Interests   []string               `json:"interests"`
	Preferences map[string]interface{} `json:"preferences"`
	Demographics map[string]interface{} `json:"demographics"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastActive  time.Time              `json:"last_active"`
	Status      string                 `json:"status"` // active, inactive, banned
	Metadata    map[string]interface{} `json:"metadata"`
}

// 用户画像模型
type UserProfile struct {
	UserID          string                 `json:"user_id"`
	FeatureVector   []float64              `json:"feature_vector"`
	PreferenceVector []float64             `json:"preference_vector"`
	BehaviorPattern map[string]interface{} `json:"behavior_pattern"`
	InterestTags    []string               `json:"interest_tags"`
	ActivityLevel   float64                `json:"activity_level"`
	EngagementScore float64                `json:"engagement_score"`
	LastUpdated     time.Time              `json:"last_updated"`
	Version         string                 `json:"version"`
	Metadata        map[string]interface{} `json:"metadata"`
}