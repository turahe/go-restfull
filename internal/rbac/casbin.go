package rbac

import (
	"errors"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type Enforcer struct {
	*casbin.Enforcer
}

func NewEnforcer(db *gorm.DB, modelPath string) (*Enforcer, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, err
	}
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", "casbin_rules")
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return &Enforcer{Enforcer: e}, nil
}

