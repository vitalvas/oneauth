package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubico"
)

func (s *Server) runHTTPServer(_ *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)

	if s.yubico == nil {
		yAuth, err := yubico.NewYubiAuth(s.config.Yubico.ClientID, s.config.Yubico.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to create YubiAuth: %w", err)
		}

		s.yubico = yAuth
	}

	r := gin.Default()

	r.GET("/-/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	{
		wellKnown := r.Group("/.well-known")
		wellKnown.GET("/security.txt", s.wellKnownSecurityTxt)
	}

	{
		v1 := r.Group("/api/v1")
		v1.GET("/yubikey/otp/verify", s.yubikeyOTPVerify)
		v1.POST("/yubikey/otp/verify", s.yubikeyOTPVerify)
	}

	return r.Run()
}
