package models

import (
	"time"
)

// 物品数据模型
type ItemData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category"`
	Brand       string                 `json:"brand"`
	Price       float64                `json:"price"`
	Currency    string                 `json:"currency"`
	Images      []string               `json:"images"`
	Tags        []string               `json:"tags"`
	Attributes  map[string]interface{} `json:"attributes"`
	Features    []float64              `json:"features"`
	QualityScore float64               `json:"quality_score"`
	Popularity  int                    `json:"popularity"`
	Rating      float64                `json:"rating"`
	RatingCount int                    `json:"rating_count"`
	Stock       int                    `json:"stock"`
	Availability string                `json:"availability"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiryDate  *time.Time             `json:"expiry_date"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 物品特征模型
type ItemFeatures struct {
	ItemID      string                 `json:"item_id"`
	CategoryVector []float64           `json:"category_vector"`
	TextFeatures   []float64           `json:"text_features"`
	ImageFeatures  []float64           `json:"image_features"`
	PriceFeature   float64              `json:"price_feature"`
	QualityFeature float64              `json:"quality_feature"`
	PopularityFeature float64           `json:"popularity_feature"`
	TemporalFeature   float64           `json:"temporal_feature"`
	FeatureVector     []float64         `json:"feature_vector"`
	FeatureVersion    string            `json:"feature_version"`
	LastUpdated       time.Time         `json:"last_updated"`
	Metadata          map[string]interface{} `json:"metadata"`
}