package hnprocessor

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/Baraahesham/hn-processor/internal/config"
	"github.com/Baraahesham/hn-processor/internal/db"
	"github.com/Baraahesham/hn-processor/internal/models"
	natsProvider "github.com/Baraahesham/hn-processor/internal/nats"
	"github.com/alitto/pond"
	"github.com/spf13/viper"

	"github.com/nats-io/nats.go"

	"github.com/rs/zerolog"
)

type HnProcessorClient struct {
	ctx          context.Context
	logger       *zerolog.Logger
	dbClient     *db.DBClient
	brandsToScan map[string]struct{}
	natsClient   *natsProvider.NatsClient
	workerPool   *pond.WorkerPool
}

type NewHnProcessorParams struct {
	Ctx        context.Context
	Logger     *zerolog.Logger
	DbClient   *db.DBClient
	Brands     []string // list of brands/keywords to detect
	NatsClient *natsProvider.NatsClient
}

var (
	wordSplitter = regexp.MustCompile(`[^\w']+`) // splits on spaces, punctuation
)

func NewHnProcessor(params NewHnProcessorParams) *HnProcessorClient {
	logger := params.Logger.With().Str("component", "HnProcessorClient").Logger()
	// build a map for brand lookup, make it lowercase for case-insensitive matching
	brandMap := make(map[string]struct{}, len(params.Brands))
	for _, b := range params.Brands {
		brandMap[strings.ToLower(b)] = struct{}{}
	}
	pool := pond.New(
		viper.GetInt(config.MaxWorkers),
		viper.GetInt(config.MaxCapacity),
		pond.Context(params.Ctx),
		pond.Strategy(pond.Balanced()),
	)
	return &HnProcessorClient{
		ctx:          params.Ctx,
		logger:       &logger,
		dbClient:     params.DbClient,
		brandsToScan: brandMap,
		natsClient:   params.NatsClient,
		workerPool:   pool,
	}
}
func (client *HnProcessorClient) Init() {
	go client.listenToHnTopStories()
}
func (client *HnProcessorClient) listenToHnTopStories() {
	client.logger.Info().Msg("Started Hacker news Top stories update event")

	err := client.natsClient.Listener(client.ctx, client.logger, string(config.NatsTopStoriesSubject), client.fanOutEvents)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to listen to Hacker news Top stories update event")
	}
}
func (client *HnProcessorClient) fanOutEvents(msg *nats.Msg) {
	client.logger.Info().Msg("Received message from NATS")
	storyEvent := models.StoryEvent{}
	err := json.Unmarshal(msg.Data, &storyEvent)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to unmarshal story event")
		return
	}
	client.workerPool.Submit(func() {
		err := client.processStoryEvent(&storyEvent)
		if err != nil {
			client.logger.Error().Err(err).Msg("Failed to process story event")
		}
	})

}
func (client *HnProcessorClient) processStoryEvent(storyEvent *models.StoryEvent) error {

	//remove all non-word characters and split the title into words
	words := wordSplitter.Split(strings.ToLower(storyEvent.Title), -1)
	var mentions []models.BrandMentionUpdateRequest

	for _, word := range words {
		word = normalizeWord(word)
		client.logger.Info().Str("word", word).Msg("Checking for brand mention")
		if _, exists := client.brandsToScan[word]; exists {
			mentions = append(mentions, models.BrandMentionUpdateRequest{
				Brand: word,
				HnID:  storyEvent.HnID,
			})
		}
	}
	if len(mentions) == 0 {
		client.logger.Info().Msg("No brand mentions found in story event")
		return nil
	}
	// Insert brand mentions into the database
	for _, mention := range mentions {
		dbModel := client.mapToBrandMentionDbModel(mention)
		err := client.dbClient.InsertBrandMention(dbModel)
		if err != nil {
			client.logger.Error().Err(err).Any("BrandMention", mention).Msg("Failed to insert brand mention into database")
			continue
		}
	}
	client.logger.Info().Msg("Brand mention inserted into database")
	return nil
}

func normalizeWord(word string) string {
	word = strings.ToLower(word)

	// Remove possessive suffix
	if strings.HasSuffix(word, "'s") {
		word = strings.TrimSuffix(word, "'s")
	}

	// Trim trailing punctuation
	word = strings.Trim(word, ".,!?\"'")

	return word
}
func (client *HnProcessorClient) mapToBrandMentionDbModel(brandMention models.BrandMentionUpdateRequest) models.BrandMentionDbModel {
	return models.BrandMentionDbModel{
		Brand: brandMention.Brand,
		HnID:  brandMention.HnID,
	}
}
