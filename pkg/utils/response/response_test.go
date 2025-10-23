package response_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	domErr "github.com/zercle/zercle-go-template/internal/core/domain/errors"
	"github.com/zercle/zercle-go-template/pkg/utils/response"
)

func TestResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, fiber.Map{"foo": "bar"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestResponse_HandleError(t *testing.T) {
	app := fiber.New()

	app.Get("/notfound", func(c *fiber.Ctx) error {
		return response.HandleError(c, domErr.ErrNotFound)
	})

	app.Get("/unknown", func(c *fiber.Ctx) error {
		return response.HandleError(c, errors.New("unknown error"))
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notfound", nil)
		resp, _ := app.Test(req)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("Unknown", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/unknown", nil)
		resp, _ := app.Test(req)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		}
	})
}
