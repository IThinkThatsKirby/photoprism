package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchFaces finds and returns faces as JSON.
//
// GET /api/v1/faces
func SearchFaces(router *gin.RouterGroup) {
	router.GET("/faces", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSubjects, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.FaceSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Faces(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
}

// GetFace returns a face as JSON.
//
// GET /api/v1/faces/:id
func GetFace(router *gin.RouterGroup) {
	router.GET("/faces/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSubjects, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		f := form.FaceSearch{ID: c.Param("id"), Markers: true}

		if results, err := search.Faces(f); err != nil || len(results) < 1 {
			Abort(c, http.StatusNotFound, i18n.ErrFaceNotFound)
			return
		} else {
			c.JSON(http.StatusOK, results[0])
		}
	})
}

// UpdateFace updates face properties.
//
// PUT /api/v1/faces/:id
func UpdateFace(router *gin.RouterGroup) {
	router.PUT("/faces/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSubjects, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.Face

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		faceId := c.Param("id")
		m := entity.FindFace(faceId)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrFaceNotFound)
			return
		}

		if err := m.SetSubjectUID(f.SubjUID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		event.SuccessMsg(i18n.MsgPersonSaved)

		c.JSON(http.StatusOK, m)
	})
}
