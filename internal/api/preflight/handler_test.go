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

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
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
			Stats: preflight.ResourcePreflight{
				Combined: preflight.PreflightStats{
					MetadataCached:   2,
					MetadataUpstream: 0,
					SubtitleCached:   0,
					SubtitleUpstream: 0,
				},
			},
		}, nil)

		err := h.Check(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp PreflightResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 1, resp.UniqueVideoCount)
		assert.Equal(t, 2, resp.Combined.MetadataCached)
		assert.Equal(t, 0, resp.Combined.MetadataUpstream)
		assert.Equal(t, 0, resp.Combined.SubtitleCached)
		assert.Equal(t, 0, resp.Combined.SubtitleUpstream)
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

	t.Run("Full integration test path", func(t *testing.T) {
		// 1. Setup Test Double for ytdlp.Provider
		// We'll create a simple mock for the Provider interface used by metadata.Service
		mockYtdlp := new(MockYtdlpProvider) // This will be defined below

		// 2. Setup Real metadata.Service
		metaSvc := metadata.NewService(nil, mockYtdlp)

		// 3. Setup Real preflight.Service
		// preflight.Service requires a cache.Manager and logger. We'll use a nil one for this test
		// as cache behavior is not part of this specific integration path test.
		preflightSvc := preflight.NewService(metaSvc, nil, nil)

		// 4. Setup Real Handler
		h := NewHandler(preflightSvc)

		// 5. Setup Request
		reqBody := `{"inputs": ["https://youtube.com/playlist?list=pl1"]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		// 6. Setup Mock Expectations
		mockYtdlp.On("GetPlaylistMetadata", mock.Anything, "pl1", mock.Anything).Run(func(args mock.Arguments) {
			st := args.Get(2).(*stats.MetadataStats)
			st.Attempts = 2
			st.Successes = 2
		}).Return(youtube.Playlist{ID: "pl1", ItemCount: 2}, nil)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// 7. Execute
		err := h.Check(c)

		// 8. Verify
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp PreflightResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 0, resp.UniqueVideoCount)
		assert.Equal(t, 2, resp.TotalResourceCount)
		assert.Equal(t, 3, resp.Combined.MetadataCached)
		assert.Equal(t, 0, resp.Combined.MetadataUpstream)
		mockYtdlp.AssertExpectations(t)
	})

	t.Run("Full integration test path (corrected)", func(t *testing.T) {
		// 1. Setup Test Double for ytdlp.Provider
		mockYtdlp := new(MockYtdlpProvider)

		// 2. Setup Real metadata.Service
		metaSvc := metadata.NewService(nil, mockYtdlp)

		// 3. Setup Real preflight.Service
		preflightSvc := preflight.NewService(metaSvc, nil, nil)

		// 4. Setup Real Handler
		h := NewHandler(preflightSvc)

		// 5. Setup Request
		reqBody := `{"inputs": ["https://youtube.com/playlist?list=pl1"]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		// 6. Setup Mock Expectations
		mockYtdlp.On("GetPlaylistMetadata", mock.Anything, "pl1", mock.Anything).Run(func(args mock.Arguments) {
			st := args.Get(2).(*stats.MetadataStats)
			st.Attempts = 1
			st.Successes = 1
		}).Return(youtube.Playlist{ID: "pl1", ItemCount: 2}, nil)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// 7. Execute
		err := h.Check(c)

		// 8. Verify
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp PreflightResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 0, resp.UniqueVideoCount)
		assert.Equal(t, 2, resp.TotalResourceCount)
		assert.Equal(t, 2, resp.Combined.MetadataCached)
		assert.Equal(t, 0, resp.Combined.MetadataUpstream)
		mockYtdlp.AssertExpectations(t)
	})
}

// MockYtdlpProvider is a mock of the metadata.Provider interface
type MockYtdlpProvider struct {
	mock.Mock
}

func (m *MockYtdlpProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	args := m.Called(ctx, playlistID, st)
	return args.Get(0).(youtube.Playlist), args.Error(1)
}

func (m *MockYtdlpProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	args := m.Called(ctx, playlistID, limit, st)
	return args.Get(0).([]youtube.PlaylistItem), args.Error(1)
}

func (m *MockYtdlpProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	args := m.Called(ctx, videoID, st)
	return args.Get(0).(youtube.Video), args.Error(1)
}
