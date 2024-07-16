package apiv1

import "github.com/gin-gonic/gin"

func (a *APIV1) createNoteHandler(c *gin.Context) {}

func (a *APIV1) getNoteBySlugHandler(c *gin.Context) {
	_ = c.Param("slug")
}
