package setting

import "webapi/internal/repository"

type SettingApp interface {
}

type settingApp struct {
	Repo *repository.Repository
}

func NewSettingApp(repo *repository.Repository) SettingApp {
	return &settingApp{
		Repo: repo,
	}
}

func (s *settingApp) Set(key, value string) {

}
