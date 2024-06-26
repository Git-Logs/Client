package state

import (
	"context"

	"github.com/git-logs/client/webserver/config"
	"github.com/git-logs/client/webserver/mapofmu"

	"github.com/bwmarrin/discordgo"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

var (
	Json       = jsoniter.ConfigCompatibleWithStandardLibrary
	Discord    *discordgo.Session
	Pool       *pgxpool.Pool
	Context    = context.Background()
	Validator  = validator.New()
	Logger     *zap.Logger
	Config     *config.Config
	MapMutex   *mapofmu.M[string]
	IsEmbedded bool // Whether gitlogs is being embedded into another service
)

const (
	configFile = "api-config.yaml"
)
