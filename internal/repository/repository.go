package repository

import (
	"webapi/internal/db/pgx"
	"webapi/internal/db/rdb"
)

type Repository struct {
	User     UserRepository
	Job      JobRepository
	Media    MediaRepository
	Setting  SettingRepository
	Tag      TagRepository
	Post     PostRepository
	Taxonomy TaxonomyRepository
}

func NewRepository() *Repository {
	pgxPool := pgx.GetPgxPool()
	redisClient := rdb.GetRedisClient()

	return &Repository{
		User:     NewUserRepository(pgxPool, redisClient),
		Job:      NewJobRepository(pgxPool),
		Media:    NewMediaRepository(pgxPool, redisClient),
		Setting:  NewSettingRepository(pgxPool, redisClient),
		Tag:      NewTagRepository(pgxPool, redisClient),
		Post:     NewPostRepository(pgxPool),
		Taxonomy: NewTaxonomyRepository(pgxPool, redisClient),
	}
}
