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

		authorized := auth.Group("/", a.authorizedMiddleware)
		{
			authorized.POST("/logout", a.logOutHandler)
			authorized.POST("/change-password", a.changePasswordHandler)
		}

		oauth := r.Group("/oauth")
		{
			oauth.GET("/google", a.googleLoginHandler)
			oauth.GET("/google/callback", a.googleCallbackHandler)

			oauth.GET("/github", a.githubLoginHandler)
			oauth.GET("/github/callback", a.githubCallbackHandler)
		}
	}

	note := r.Group("/note", a.couldBeAuthorizedMiddleware)
	{
		note.GET("/:slug", a.getNoteBySlugHandler)
		note.POST("", a.createNoteHandler)
	}
}
