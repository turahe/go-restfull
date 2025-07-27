package cache

const (
	// Legacy keys
	KEY_THKCORE_ACCESS_TOKEN        string = "thkcore_access_token"
	KEY_MOBILE_BACKEND_ACCESS_TOKEN string = "mobile_backend_access_token"

	// User cache keys
	KEY_USER_BY_ID    string = "user:by_id:%s"
	KEY_USER_BY_EMAIL string = "user:by_email:%s"
	KEY_USER_BY_PHONE string = "user:by_phone:%s"
	KEY_USERS_ALL     string = "users:all"
	KEY_USERS_PAGE    string = "users:page:%d:limit:%d"
	KEY_USER_COUNT    string = "users:count"

	// Post cache keys
	KEY_POST_BY_ID    string = "post:by_id:%s"
	KEY_POST_BY_SLUG  string = "post:by_slug:%s"
	KEY_POSTS_ALL     string = "posts:all"
	KEY_POSTS_PAGE    string = "posts:page:%d:limit:%d"
	KEY_POSTS_BY_USER string = "posts:by_user:%s"
	KEY_POST_COUNT    string = "posts:count"

	// Media cache keys
	KEY_MEDIA_BY_ID       string = "media:by_id:%s"
	KEY_MEDIA_BY_HASH     string = "media:by_hash:%s"
	KEY_MEDIA_BY_FILENAME string = "media:by_filename:%s"
	KEY_MEDIA_ALL         string = "media:all"
	KEY_MEDIA_PAGE        string = "media:page:%d:limit:%d"
	KEY_MEDIA_BY_PARENT   string = "media:by_parent:%s"
	KEY_MEDIA_COUNT       string = "media:count"

	// Tag cache keys
	KEY_TAG_BY_ID      string = "tag:by_id:%s"
	KEY_TAG_BY_SLUG    string = "tag:by_slug:%s"
	KEY_TAGS_ALL       string = "tags:all"
	KEY_TAGS_PAGE      string = "tags:page:%d:limit:%d"
	KEY_TAGS_BY_ENTITY string = "tags:by_entity:%s:%s"
	KEY_TAG_COUNT      string = "tags:count"

	// Comment cache keys
	KEY_COMMENT_BY_ID    string = "comment:by_id:%s"
	KEY_COMMENTS_ALL     string = "comments:all"
	KEY_COMMENTS_PAGE    string = "comments:page:%d:limit:%d"
	KEY_COMMENTS_BY_POST string = "comments:by_post:%s"
	KEY_COMMENT_COUNT    string = "comments:count"

	// Role cache keys
	KEY_ROLE_BY_ID   string = "role:by_id:%s"
	KEY_ROLE_BY_NAME string = "role:by_name:%s"
	KEY_ROLES_ALL    string = "roles:all"
	KEY_ROLES_PAGE   string = "roles:page:%d:limit:%d"
	KEY_ROLE_COUNT   string = "roles:count"

	// UserRole cache keys
	KEY_USER_ROLES_BY_USER string = "user_roles:by_user:%s"
	KEY_USER_ROLES_BY_ROLE string = "user_roles:by_role:%s"
	KEY_USER_ROLE_EXISTS   string = "user_role:exists:%s:%s"

	// Menu cache keys
	KEY_MENU_BY_ID      string = "menu:by_id:%s"
	KEY_MENU_BY_SLUG    string = "menu:by_slug:%s"
	KEY_MENUS_ALL       string = "menus:all"
	KEY_MENUS_PAGE      string = "menus:page:%d:limit:%d"
	KEY_MENUS_HIERARCHY string = "menus:hierarchy"
	KEY_MENU_COUNT      string = "menus:count"

	// MenuRole cache keys
	KEY_MENU_ROLES_BY_MENU string = "menu_roles:by_menu:%s"
	KEY_MENU_ROLES_BY_ROLE string = "menu_roles:by_role:%s"
	KEY_MENU_ROLE_EXISTS   string = "menu_role:exists:%s:%s"

	// Taxonomy cache keys
	KEY_TAXONOMY_BY_ID       string = "taxonomy:by_id:%s"
	KEY_TAXONOMY_BY_SLUG     string = "taxonomy:by_slug:%s"
	KEY_TAXONOMIES_ALL       string = "taxonomies:all"
	KEY_TAXONOMIES_PAGE      string = "taxonomies:page:%d:limit:%d"
	KEY_TAXONOMIES_HIERARCHY string = "taxonomies:hierarchy"
	KEY_TAXONOMY_COUNT       string = "taxonomies:count"

	// Setting cache keys
	KEY_SETTING_BY_KEY string = "setting:by_key:%s"
	KEY_SETTINGS_ALL   string = "settings:all"

	// Job cache keys
	KEY_JOB_BY_ID  string = "job:by_id:%s"
	KEY_JOBS_ALL   string = "jobs:all"
	KEY_JOBS_PAGE  string = "jobs:page:%d:limit:%d"
	KEY_JOBS_QUEUE string = "jobs:queue:%s"
	KEY_JOB_COUNT  string = "jobs:count"

	// Cache invalidation patterns
	PATTERN_USER_CACHE     string = "user:*"
	PATTERN_POST_CACHE     string = "post:*"
	PATTERN_MEDIA_CACHE    string = "media:*"
	PATTERN_TAG_CACHE      string = "tag:*"
	PATTERN_COMMENT_CACHE  string = "comment:*"
	PATTERN_ROLE_CACHE     string = "role:*"
	PATTERN_MENU_CACHE     string = "menu:*"
	PATTERN_TAXONOMY_CACHE string = "taxonomy:*"
	PATTERN_SETTING_CACHE  string = "setting:*"
	PATTERN_JOB_CACHE      string = "job:*"
)
