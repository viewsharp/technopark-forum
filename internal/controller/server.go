package controller

import (
	"github.com/viewsharp/technopark-forum/internal/repository"
)

// Server implements api.ServerInterface
type Server struct {
	sb *repository.RepositorySet
}

func NewServer(repositorySet *repository.RepositorySet) *Server {
	return &Server{sb: repositorySet}
}
