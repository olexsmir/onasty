package apiv1

import (
	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
)

type APIV1 struct {
	usersrv usersrv.UserServicer
	notesrv notesrv.NoteServicer
}

func NewAPIV1(
	us usersrv.UserServicer,
	ns notesrv.NoteServicer,
) *APIV1 {
	return &APIV1{
		usersrv: us,
		notesrv: ns,
	}
}

func (a *APIV1) Routes(r *gin.RouterGroup) {
	r.Use(a.metricsMiddleware)
	auth := r.Group("/auth")
	{
		auth.POST("/signup", a.signUpHandler)
		auth.POST("/signin", a.signInHandler)
		auth.POST("/refresh-tokens", a.refreshTokensHandler)
		auth.GET("/verify/:token", a.verifyHandler)
		auth.POST("/resend-verification-email", a.resendVerificationEmailHandler)
		auth.POST("/reset-password", a.requestResetPasswordHandler)
		auth.POST("/reset-password/:token", a.resetPasswordHandler)

		oauth := r.Group("/oauth")
		{
			oauth.GET("/:provider", a.oauthLoginHandler)
			oauth.GET("/:provider/callback", a.oauthCallbackHandler)
		}

		authorized := auth.Group("/", a.authorizedMiddleware)
		{
			authorized.POST("/logout", a.logOutHandler)
			authorized.POST("/change-password", a.changePasswordHandler)
		}
	}

	note := r.Group("/note")
	{
		note.GET("/:slug", a.getNoteBySlugHandler)

		possiblyAuthorized := note.Group("", a.couldBeAuthorizedMiddleware)
		{
			possiblyAuthorized.POST("", a.createNoteHandler)
		}
	}
}
