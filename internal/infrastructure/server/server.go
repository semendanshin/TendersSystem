package server

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/infrastructure/server/handlers"
	"tenderSystem/internal/infrastructure/server/middleware"
)

type Server struct {
	tenderUseCase abstraction.TenderUseCaseInterface
	bidsUseCase   abstraction.BidUseCaseInterface

	e    *echo.Echo
	host string
	port string
}

func NewServer(
	tenderUseCase abstraction.TenderUseCaseInterface, bidsUseCase abstraction.BidUseCaseInterface,
	host string, port string,
) *Server {
	return &Server{
		tenderUseCase: tenderUseCase,
		bidsUseCase:   bidsUseCase,
		e:             echo.New(),
		host:          host,
		port:          port,
	}
}

func (s *Server) Start() error {
	//add /api prefix

	g := s.e.Group("/api")

	pingHandler := handlers.NewPingHandler()
	pingHandler.Register(g)

	tenderHandler := handlers.NewTenderHandler(s.tenderUseCase)
	tenderHandler.Register(g)

	bidHandler := handlers.NewBidHandler(s.bidsUseCase)
	bidHandler.Register(g)

	s.e.Use(echoMiddleware.Logger())
	s.e.Use(middleware.NewErrorMiddleware())
	s.e.Use(echoMiddleware.Recover())

	return s.e.Start(s.host + ":" + s.port)
}
