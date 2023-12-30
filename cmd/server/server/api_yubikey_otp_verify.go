package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type YubikeyOTPVerifyRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	OTP      string `json:"otp" form:"otp" binding:"required"`
}

func (s *Server) yubikeyOTPVerify(ginCtx *gin.Context) {
	var request YubikeyOTPVerifyRequest

	switch ginCtx.Request.Method {
	case http.MethodGet:
		if err := ginCtx.ShouldBindQuery(&request); err != nil {
			ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	case http.MethodPost:
		if err := ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	default:
		ginCtx.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}

	valid, err := s.yubico.Verify(request.OTP)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ginCtx.JSON(http.StatusOK, valid)
}
