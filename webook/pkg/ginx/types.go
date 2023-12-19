package ginx

import "github.com/gin-gonic/gin"

type Server struct {
	*gin.Engine
	Addr string
}

func (s *Server) Start() error {
	return s.Engine.Run(s.Addr)
}
