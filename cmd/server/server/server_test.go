package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerStruct(t *testing.T) {
	t.Run("ZeroValue", func(t *testing.T) {
		srv := Server{}
		assert.Nil(t, srv.config)
		assert.Nil(t, srv.yubico)
	})

	t.Run("WithConfig", func(t *testing.T) {
		cfg := &Config{
			Yubico: ConfigYubico{
				ClientID:     42,
				ClientSecret: "secret",
			},
		}
		srv := Server{config: cfg}
		assert.NotNil(t, srv.config)
		assert.Equal(t, 42, srv.config.Yubico.ClientID)
	})
}
