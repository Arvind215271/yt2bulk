package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"

	// config
	"github.com/Arvind215271/yt2bulk/internal/config"

	// logger
	"github.com/Arvind215271/yt2bulk/internal/logger"

	// api
	"github.com/Arvind215271/yt2bulk/internal/api"
	"github.com/Arvind215271/yt2bulk/internal/api/export"
	apiPlaylist "github.com/Arvind215271/yt2bulk/internal/api/playlist"
	apiPreflight "github.com/Arvind215271/yt2bulk/internal/api/preflight"
	apiSubtitle "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
	apiTranscript "github.com/Arvind215271/yt2bulk/internal/api/transcript"

	// cache
	"github.com/Arvind215271/yt2bulk/internal/cache"

	// youtube
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	youtubeapi "github.com/Arvind215271/yt2bulk/internal/youtube/metadata/youtube_api"
	ytdlpmetadata "github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp/python"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	ytSubtitle "github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"

	// transcript
	ytTranscript "github.com/Arvind215271/yt2bulk/internal/youtube/transcript"

	// signal
	ytSignal "github.com/Arvind215271/yt2bulk/internal/youtube/signal"

	// export
	exportservice "github.com/Arvind215271/yt2bulk/internal/youtube/export"

	// description
	"github.com/Arvind215271/yt2bulk/internal/youtube/description"

	// debug
	"github.com/Arvind215271/yt2bulk/debug"

	//pipeline
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"

	//resource
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
)

func main() {

	// get the config
	config := config.Load()

	// get the logger
	logger := logger.New(
		config.Env,
	)

	logger.Info(
		config.Env,
	)

	// get the error handler
	errorHandler := api.NewErrorHandler(
		logger,
	)

	// create echo instance
	e := echo.New()

	// CORS middleware with secure specific origins
	frontendOrigin := os.Getenv("FRONTEND_URL")
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			origin := c.Request().Header.Get("Origin")
			if origin == "http://localhost:3000" ||
				origin == "http://127.0.0.1:3000" ||
				(frontendOrigin != "" && origin == frontendOrigin) {
				c.Response().Header().Set("Access-Control-Allow-Origin", origin)
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusOK)
			}
			return next(c)
		}
	})

	// ping route
	e.GET(
		"/ping",
		ping,
	)

	// context to be used by youtube API
	ctx := context.Background()

	// youtube

	youtubeClient, err := youtubeapi.NewClient(
		ctx,
		config.YouTubeAPIKey,
		logger,
	)
	if err != nil {
		logger.Warn(
			"failed to create youtube client",
			"error",
			err,
		)
		youtubeClient = nil
	}

	youtubeProvider := youtubeapi.NewProvider(
		youtubeClient,
		logger,
	)

	// cache manager
	cacheManager := cache.NewManager(config.YtdlpCache, logger)

	// cleanup.
	cacheManager.Cleanup()

	// start periodic cleanup loop
	cacheManager.StartCleanupLoop(ctx, config.YtdlpCache.CleanupInterval)

	pythonPool, err := python.NewSinglePool(config.PythonWorkers, logger, cacheManager)
	pythonPool.WarmUp(ctx)

	if err != nil {
		logger.Fatal("failed to create python pool", "error", err)
	}

	ytdlpMetadataClient := ytdlpmetadata.NewClient(pythonPool, logger)

	ytdlpMetadataProvider := ytdlpmetadata.NewProvider(ytdlpMetadataClient, logger)

	// run cleanup on startup
	if err := ytdlpMetadataClient.Cleanup(); err != nil {
		logger.Error("failed to perform ytdlp cache cleanup", "error", err)
	} else {
		logger.Info("ytdlp cache cleanup completed successfully")
	}

	youtubeService := metadata.NewService(
		youtubeProvider,
		ytdlpMetadataProvider,
	)

	subtitleService := ytSubtitle.NewSubtitleService(cacheManager, logger, pythonPool)
	preflightService := preflight.NewService(youtubeService, cacheManager, logger, config.MaxExportLimit)
	transcriptService := ytTranscript.NewService()
	signalService := ytSignal.NewSignalService()

	// description
	descriptionService := description.NewService()

	// pipeline
	pipelineService := pipeline.NewService(youtubeService, descriptionService, subtitleService, transcriptService, signalService, logger, 2*config.PythonWorkers)

	// export service
	exportService := exportservice.NewService()

	exportService.RegisterExporter(
		exportservice.FormatJSON,
		&exportservice.JSONExporter{Pretty: true},
	)
	exportService.RegisterExporter(
		exportservice.FormatCSV,
		&exportservice.CSVExporter{},
	)
	exportService.RegisterExporter(
		exportservice.FormatMarkdown,
		&exportservice.MarkdownExporter{},
	)
	exportService.RegisterExporter(
		exportservice.FormatExcel,
		&exportservice.ExcelExporter{},
	)
	exportService.RegisterExporter(
		exportservice.FormatYAML,
		&exportservice.YAMLExporter{},
	)
	exportService.RegisterExporter(
		exportservice.FormatXML,
		&exportservice.XMLExporter{},
	)

	// resource service
	resourceService := resource.NewService(youtubeService, pipelineService, logger, 2*config.PythonWorkers)

	// handlers & routes
	exportHandler := export.NewHandler(resourceService, exportService, preflightService, config)
	export.RegisterRoutes(e.Group("/export"), exportHandler)

	subtitleHandler := apiSubtitle.NewHandler(resourceService, subtitleService, exportService, preflightService)
	apiSubtitle.RegisterSubtitleRoutes(e.Group("/subtitle"), subtitleHandler)

	transcriptHandler := apiTranscript.NewHandler(resourceService, preflightService)
	e.POST("/transcript", transcriptHandler.GetTranscript)

	playlistHandler := apiPlaylist.NewHandler(youtubeService, preflightService)
	apiPlaylist.RegisterPlaylistRoutes(e.Group("/playlist"), playlistHandler)

	preflightHandler := apiPreflight.NewHandler(preflightService)
	apiPreflight.RegisterRoutes(e.Group("/preflight"), preflightHandler)

	// only run debug in development
	if config.Env == "dev" {

		debug.DebugInput(
			ctx,
			logger,
			youtubeService,
			transcriptService,
			exportService,
		)
	}

	e.HTTPErrorHandler = errorHandler.Handle

	// start the server

	if err := e.Start(":" + config.Port); err != nil {
		fmt.Println(
			"FAILED TO START THE SERVER",
			"ERROR:",
			err,
		)
	}
}

func ping(
	c *echo.Context,
) error {

	return c.JSON(
		http.StatusOK,
		"pong",
	)
}
