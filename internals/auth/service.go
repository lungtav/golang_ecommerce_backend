package auth

import "github.com/lungtav/golang_ecommerce_backend/internals/db"

type Service interface {
	RegisterUser() error
}

type svc struct {
	repo *db.Queries
}

func AuthService (repo *db.Queries ) *svc {
	return &svc{
		repo: repo,
	}
}

func (s *svc) RegisterUser() error {
	return nil
}
