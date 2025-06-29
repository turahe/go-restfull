package setting

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/dto"
	"webapi/internal/http/requests"
	"webapi/internal/repository"
)

type SettingApp interface {
	CreateSetting(ctx context.Context, req requests.CreateSettingRequest) (*dto.SettingDTO, error)
	GetSettingByKey(ctx context.Context, req requests.GetSettingByKeyRequest) (*dto.SettingDTO, error)
	GetAllSettings(ctx context.Context) ([]dto.SettingDTO, error)
	UpdateSetting(ctx context.Context, req requests.UpdateSettingRequest) (*dto.SettingDTO, error)
	DeleteSetting(ctx context.Context, key string) error
}

type settingApp struct {
	Repo *repository.Repository
}

func NewSettingApp(repo *repository.Repository) SettingApp {
	return &settingApp{
		Repo: repo,
	}
}

func (s *settingApp) CreateSetting(ctx context.Context, req requests.CreateSettingRequest) (*dto.SettingDTO, error) {
	setting := model.Setting{
		ModelType: req.ModelType,
		ModelId:   req.ModelId,
		Key:       req.Key,
		Value:     req.Value,
	}

	err := s.Repo.Setting.SetModelSetting(ctx, setting)
	if err != nil {
		return nil, err
	}

	// Get the created setting
	createdSetting, err := s.Repo.Setting.GetSettingByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return &dto.SettingDTO{
		ID:        createdSetting.ID,
		ModelType: createdSetting.ModelType,
		ModelId:   createdSetting.ModelId,
		Key:       createdSetting.Key,
		Value:     createdSetting.Value,
		CreatedAt: createdSetting.CreatedAt,
		UpdatedAt: createdSetting.UpdatedAt,
	}, nil
}

func (s *settingApp) GetSettingByKey(ctx context.Context, req requests.GetSettingByKeyRequest) (*dto.SettingDTO, error) {
	setting, err := s.Repo.Setting.GetSettingByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return &dto.SettingDTO{
		ID:        setting.ID,
		ModelType: setting.ModelType,
		ModelId:   setting.ModelId,
		Key:       setting.Key,
		Value:     setting.Value,
		CreatedAt: setting.CreatedAt,
		UpdatedAt: setting.UpdatedAt,
	}, nil
}

func (s *settingApp) GetAllSettings(ctx context.Context) ([]dto.SettingDTO, error) {
	// This would need to be implemented in the repository
	// For now, we'll return an empty slice
	return []dto.SettingDTO{}, nil
}

func (s *settingApp) UpdateSetting(ctx context.Context, req requests.UpdateSettingRequest) (*dto.SettingDTO, error) {
	updatedSetting, err := s.Repo.Setting.UpdateSetting(ctx, req.Key, req.Value)
	if err != nil {
		return nil, err
	}

	return &dto.SettingDTO{
		ID:        updatedSetting.ID,
		ModelType: updatedSetting.ModelType,
		ModelId:   updatedSetting.ModelId,
		Key:       updatedSetting.Key,
		Value:     updatedSetting.Value,
		CreatedAt: updatedSetting.CreatedAt,
		UpdatedAt: updatedSetting.UpdatedAt,
	}, nil
}

func (s *settingApp) DeleteSetting(ctx context.Context, key string) error {
	setting := model.Setting{Key: key}
	return s.Repo.Setting.DeleteSetting(ctx, setting)
}
