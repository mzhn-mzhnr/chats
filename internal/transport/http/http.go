package http

import (
	"context"
	"fmt"
	"log/slog"
	"mzhn/chats/internal/config"
	"mzhn/chats/internal/services/authservice"
	"mzhn/chats/internal/services/chatservice"
	"mzhn/chats/internal/transport/http/handlers"
	"mzhn/chats/internal/transport/http/middleware"
	"mzhn/chats/pkg/sl"
	"strings"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	eswag "github.com/swaggo/echo-swagger"

	_ "mzhn/chats/docs"
)

type Server struct {
	*echo.Echo

	cfg    *config.Config
	logger *slog.Logger

	cs *chatservice.Service
	as *authservice.Service
}

func New(cfg *config.Config, cs *chatservice.Service, as *authservice.Service) *Server {
	return &Server{
		Echo:   echo.New(),
		logger: slog.Default().With(sl.Module("http")),
		cfg:    cfg,
		cs:     cs,
		as:     as,
	}
}

func (h *Server) setup() {

	h.Use(emw.Logger())
	h.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins:     strings.Split(h.cfg.Http.Cors.AllowedOrigins, ","),
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowCredentials: true,
	}))

	h.GET("/docs/*", eswag.WrapHandler)

	authguard := middleware.AuthGuard(h.as)
	h.GET("/", handlers.GetConversations(h.cs), authguard())
	h.GET("/:id", handlers.GetConversation(h.cs), authguard())
	h.POST("/", handlers.CreateConversation(h.cs), authguard())
}

// @title			MZHN Chat API
// @version		0.1
// @description	Chat Api Service
// @contact.url http://github.com/mzhn-mzhnr/chats
// @BasePath		/conversations/
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func (h *Server) Run(ctx context.Context) error {
	h.setup()

	host := h.cfg.Http.Host
	port := h.cfg.Http.Port
	addr := fmt.Sprintf("%s:%d", host, port)
	slog.Info("running http server", slog.String("addr", addr))

	go func() {
		if err := h.Start(addr); err != nil {
			return
		}
	}()

	<-ctx.Done()
	if err := h.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("shutting down http server\n")
	return nil
}
