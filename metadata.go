package star

import (
	"time"

	"github.com/tubblelive/star/star/crypto"
)

var Version string
var StartedAt = time.Now()
var NodeID = crypto.CreateRandString(12)

func GetVersion() string {
	if Version == "" {
		return "dev"
	}

	return Version
}