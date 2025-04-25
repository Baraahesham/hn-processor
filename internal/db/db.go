package db

import (
	"github.com/Baraahesham/hn-processor/internal/models"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBClient struct {
	Gorm   *gorm.DB
	logger *zerolog.Logger
}
type NewDBClientParams struct {
	Logger *zerolog.Logger
	DbUrl  string
}

func NewClient(params NewDBClientParams) (*DBClient, error) {
	// Initialize the logger
	logger := params.Logger.With().Str("component", "DBClient").Logger()

	logger.Info().Str("db_url", params.DbUrl).Msg("Connecting to PostgreSQL")

	gormDB, err := gorm.Open(postgres.Open(params.DbUrl), &gorm.Config{})

	if err != nil {
		logger.Error().Err(err).Str("db_url", params.DbUrl).Msg("Failed to connect to PostgreSQL")
		return nil, err
	}

	err = gormDB.AutoMigrate(&models.BrandMentionDbModel{})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to auto-migrate Story model")
		return nil, err
	}

	logger.Info().Str("db_url", params.DbUrl).Msg("Successfully Connected to PostgreSQL")
	return &DBClient{
		Gorm:   gormDB,
		logger: &logger,
	}, nil
}

// InsertStory inserts a story into the database
func (client *DBClient) InsertBrandMention(brandMention models.BrandMentionDbModel) error {
	err := client.Gorm.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "brand"}, {Name: "hn_id"}},
		DoNothing: true,
	}).Create(&brandMention).Error
	if err != nil {
		client.logger.Error().Err(err).Str("brand", brandMention.Brand).Msg("Failed to insert brand into database")
		return err
	}
	client.logger.Info().Str("brand", brandMention.Brand).Msg("Story successfully inserted into database")
	return nil
}

func (client *DBClient) Close() error {
	sqlDB, err := client.Gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
