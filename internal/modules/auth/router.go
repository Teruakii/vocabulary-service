package auth

import (
	"github.com/gin-gonic/gin"

	"vocabulary-service/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authSvc Service) {
	a := rg.Group("/auth")
	{
		a.POST("/register", h.Register)
		a.POST("/login", h.Login)
		a.POST("/refresh", h.Refresh)
		a.POST("/logout", h.Logout)
		a.GET("/google", h.GoogleLogin)
		a.GET("/google/callback", h.GoogleCallback)
		a.GET("/me", middleware.OptionalAuth(authSvc), h.Me)
		a.PUT("/me", middleware.RequireAuth(authSvc), h.UpdateProfile)
		a.DELETE("/me", middleware.RequireAuth(authSvc), h.Delete)
		a.POST("/forgot-password", h.ForgotPassword)
		a.POST("/reset-password", h.ResetPassword)
		a.POST("/change-password", middleware.RequireAuth(authSvc), h.ChangePassword)
		a.POST("/admin/users/:id/reset-password",
			middleware.RequireAuth(authSvc), middleware.RequireAdmin(), h.AdminResetPassword)

		a.GET("/identities", middleware.RequireAuth(authSvc), h.ListIdentities)
		a.POST("/identities", middleware.RequireAuth(authSvc), h.LinkIdentity)
		a.DELETE("/identities/:id", middleware.RequireAuth(authSvc), h.UnlinkIdentity)
	}
}
