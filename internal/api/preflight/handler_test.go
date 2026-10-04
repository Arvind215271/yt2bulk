package preflight

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Arvind215271/askito/internal/youtube/preflight"
)

// MockService is a mock of the preflight.Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) Check(ctx context.Context, req preflight.Request) (preflight.Result, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(preflight.Result), args.Error(1)
}

func TestHandler_Check(t *testing.T) {
	e := echo.New()

	t.Run("Valid request returns success", func(t *testing.T) {
		mockSvc := new(MockService)
		h := NewHandler(mockSvc)

		reqBody := `{"inputs": ["https://youtube.com/watch?v=123"]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Expectation
		mockSvc.On("Check", mock.Anything, mock.Anything).Return(preflight.Result{
			Valid:              true,
			UniqueVideoCount:   1,
			TotalResourceCount: 1,
			EstimatedWork:      1,
		}, nil)

		err := h.Check(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp PreflightResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 1, resp.UniqueVideoCount)
	})

	t.Run("Invalid request returns error", func(t *testing.T) {
		mockSvc := new(MockService)
		h := NewHandler(mockSvc)

		reqBody := `{"inputs": []}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Check(c)

		// Expectation: BadRequest (handled by our error handler)
		assert.Error(t, err)
	})
}
