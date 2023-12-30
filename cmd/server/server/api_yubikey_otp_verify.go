package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitalvas/oneauth/internal/yubico"
)

type YubikeyOTPVerifyRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	OTP      string `json:"otp" form:"otp" binding:"required"`
}

type YubikeyOTPVerifyResponse struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
	Serial   int64  `json:"serial"`
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

	if valid.Status != yubico.StatusOK {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid OTP: %s", valid.Status)})
		return
	}

	ginCtx.JSON(http.StatusOK, YubikeyOTPVerifyResponse{
		Username: request.Username,
		OTP:      request.OTP,
		Serial:   valid.Serial,
	})
}
