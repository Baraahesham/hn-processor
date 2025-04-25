package models

import "github.com/nats-io/nats.go"

type StoryEvent struct {
	HnID  int64  `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}
type BrandMentionDbModel struct {
	ID    uint   `gorm:"primaryKey"`
	Brand string `gorm:"index:idx_brand_story,unique"`
	HnID  int64  `gorm:"index:idx_brand_story,unique"` // Hacker News ID (foreign ref to stories.hn_id)
}
type ListenRequest struct {
	Subject  string
	CallBack nats.MsgHandler
}
type BrandMentionUpdateRequest struct {
	Brand string
	HnID  int64
}

func (BrandMentionDbModel) TableName() string {
	return "brand_mentions"
}
