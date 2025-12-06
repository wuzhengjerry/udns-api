package conf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuzhengjerry/udns-api/conf"
)

func TestLoadConfigFromToml(t *testing.T) {
	should := require.New(t)

	err := conf.LoadConfigFromToml("../etc/keyauth.toml")
	should.NoError(err)

	t.Log(conf.C().Mongo.Endpoints)
}

func TestMongoClient(t *testing.T) {
	should := require.New(t)

	err := conf.LoadConfigFromToml("../etc/keyauth.toml")
	should.NoError(err)

	conf.C().Mongo.Client()
}
