package postgres

import "context"

type TenentService struct {
	db *DB
}

var tenentService *TenentService

func NewTenentService(db *DB) *TenentService {
	if tenentService == nil {
		tenentService = &TenentService{}
	}
	tenentService.db = db

	return tenentService
}

func (s *TenentService) Set(ctx context.Context) error {
	return s.db.setUserIDPerConnection(ctx)
}
