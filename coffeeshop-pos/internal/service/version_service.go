package service

import "coffeeshop-pos/internal/version"

type VersionService struct{}

func NewVersionService() *VersionService { return &VersionService{} }

func (s *VersionService) GetVersion() string { return version.Version }

func (s *VersionService) GetFullVersion() map[string]string {
	return map[string]string{
		"version":   version.Version,
		"commit":    version.CommitSHA,
		"buildDate": version.BuildDate,
	}
}
