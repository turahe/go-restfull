package model

// Table names for post-related persistence (queries, joins, raw SQL).
const (
	TablePosts     = "posts"
	TablePostSEO   = "post_seo"
	TablePostMedia = "post_media"
	TablePostTags  = "post_tags"
)

func (Post) TableName() string {
	return TablePosts
}

// PostSEO is 1:1 optional SEO / sharing metadata for a post.
type PostSEO struct {
	PostID uint `json:"-" gorm:"primaryKey"`

	Excerpt         string `json:"excerpt,omitempty" gorm:"type:text"`
	MetaTitle       string `json:"metaTitle,omitempty" gorm:"type:varchar(200);index"`
	MetaDescription string `json:"metaDescription,omitempty" gorm:"type:varchar(320)"`
	CanonicalURL    string `json:"canonicalUrl,omitempty" gorm:"type:varchar(512)"`
	OgImageURL      string `json:"ogImageUrl,omitempty" gorm:"type:varchar(512)"`
	RobotsMeta      string `json:"robotsMeta,omitempty" gorm:"type:varchar(100)"`
}

func (PostSEO) TableName() string {
	return TablePostSEO
}

// PostMedia is the many2many join between posts and media.
// Matches GORM's default join table for Post.Media.
type PostMedia struct {
	PostID  uint `json:"postId" gorm:"primaryKey;index"`
	MediaID uint `json:"mediaId" gorm:"primaryKey;index"`
}

func (PostMedia) TableName() string {
	return TablePostMedia
}

// PostTag is the many2many join between posts and tags.
// Matches GORM's default join table for Post.Tags.
type PostTag struct {
	PostID uint `json:"postId" gorm:"primaryKey;index"`
	TagID  uint `json:"tagId" gorm:"primaryKey;index"`
}

func (PostTag) TableName() string {
	return TablePostTags
}
