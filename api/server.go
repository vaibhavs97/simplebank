package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/vaibhavs97/simplebank/db/sqlc"
)

// Server serves HTTP requests for our banking Service.
type Server struct {
	store  db.Store    // for interacting with the db when processing API requests from clients.
	router *gin.Engine // for send each API request to the currect handler for processing
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.PATCH("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error:": err.Error()}
}
