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
	auth := r.Group("/auth")
	{
		auth.POST("/signup", a.signUpHandler)
		auth.POST("/signin", a.signInHandler)
		auth.POST("/refresh-tokens", a.refreshTokensHandler)

		authorized := auth.Group("/", a.authorizedMiddleware)
		{
			authorized.POST("/logout", a.logOutHandler)
		}
	}

	note := r.Group("/note", a.couldBeAuthorizedMiddleware)
	{
		note.GET("/:slug", a.getNoteBySlugHandler)
		note.POST("", a.createNoteHandler)
	}
}
