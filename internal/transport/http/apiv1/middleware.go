package apiv1

import "github.com/gin-gonic/gin"

func (a *APIV1) authorizedMiddleware(_ *gin.Context) {}

func (a *APIV1) couldBeAuthorizedMiddleware(_ *gin.Context) { //nolint:unused
}
