package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) wellKnownSecurityTxt(ginCtx *gin.Context) {
	body := `Contact: https://github.com/vitalvas/oneauth/issues
Contact: mailto:oneauth+security@vitalvas.dev
`
	ginCtx.String(http.StatusOK, body)
}
