package state

import (
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database"
	"github.com/tubblelive/star/star/database/model"
	"github.com/zitadel/oidc/v3/pkg/op"
)

var StarState *Server
var Database *database.Database

type Server struct {
	Config     *config.Config
	SigningKey *model.Keypair
	Provider   *op.Provider
}
