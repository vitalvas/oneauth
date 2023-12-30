package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) wellKnownSecurityTxt(ginCtx *gin.Context) {
	body := `Contact: https://github.com/vitalvas/oneauth/issues
Contact: mailto:oneauth+security@vitalvas.dev
Hiring: https://github.com/vitalvas/oneauth
`
	ginCtx.String(http.StatusOK, body)
}

func (s *Server) wellKnownOneAuth(ginCtx *gin.Context) {
	ginCtx.JSON(http.StatusOK, gin.H{})
}
