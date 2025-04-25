package nats

import (
	"context"

	"github.com/Baraahesham/hn-processor/internal/models"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type NatsClient struct {
	natsConn *nats.Conn
	logger   zerolog.Logger
}
type NewNatsClientParams struct {
	Logger  *zerolog.Logger
	NatsUrl string
}

func New(params NewNatsClientParams) (*NatsClient, error) {
	logger := params.Logger.With().Str("component", "NatsClient").Logger()
	natsConn, err := nats.Connect(params.NatsUrl)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to NATS")
		return nil, err
	}
	return &NatsClient{
		natsConn: natsConn,
		logger:   logger,
	}, nil
}
func (client *NatsClient) Listener(ctx context.Context, logger *zerolog.Logger, subject string, callback func(msg *nats.Msg)) error {

	logger.Info().
		Str("subject", subject).
		Msg("Start listening for requests")

	sub, err := client.listen(models.ListenRequest{Subject: subject, CallBack: callback})

	if err != nil {
		logger.Error().
			Err(err).
			Str("subject", subject).
			Msg("Failed to connect listener")

		return err
	}

	if sub != nil {
		defer sub.Unsubscribe()
	}

	<-ctx.Done()

	return nil
}

func (client *NatsClient) listen(request models.ListenRequest) (*nats.Subscription, error) {
	sub, err := client.natsConn.Subscribe(request.Subject, request.CallBack)

	if err != nil {
		client.logger.Error().
			Err(err).
			Str("Subject", request.Subject).
			Msg("Failed to subscribe to subject")
		return nil, err
	}
	return sub, nil
}
func (client *NatsClient) Close() {
	if client.natsConn != nil {
		client.logger.Info().Msg("Closing NATS connection")
		client.natsConn.Close()
	}
}
