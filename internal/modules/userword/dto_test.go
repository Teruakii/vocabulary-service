package userword

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

func TestReviewRequest_QualityRequired(t *testing.T) {
	for _, body := range []string{`{}`, `{"quality": null}`, `{"qualty": 4}`} {
		var req ReviewRequest
		assert.Error(t, binding.JSON.BindBody([]byte(body), &req), body)
	}
}

func TestReviewRequest_AcceptsZeroToFive(t *testing.T) {
	var req ReviewRequest
	assert.NoError(t, binding.JSON.BindBody([]byte(`{"quality": 0}`), &req))
	assert.Equal(t, 0, *req.Quality)

	assert.Error(t, binding.JSON.BindBody([]byte(`{"quality": 6}`), &ReviewRequest{}))
	assert.Error(t, binding.JSON.BindBody([]byte(`{"quality": -1}`), &ReviewRequest{}))
}
