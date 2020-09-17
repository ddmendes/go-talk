package dao

import (
	"database/sql"

	"github.com/ddmendes/go-talk/update-job/data"
	"github.com/ddmendes/go-talk/update-job/model"
	"gorm.io/gorm"
)

// RepositoryDAO is an access point to Repository persistence
type RepositoryDAO struct {
	db       *gorm.DB
	pageSize int
}

// NewRepositoryDAO creates DAO for Repository model
func NewRepositoryDAO() (*RepositoryDAO, error) {
	db, err := data.GetDB()
	if err != nil {
		return nil, err
	}

	pageSize := data.Config.PageSize

	dao := &RepositoryDAO{db, pageSize}
	return dao, nil
}

// Save writes a Repository to database
func (dao *RepositoryDAO) Save(repo *model.Repository) {
	dao.db.Save(repo)
}

// GetAllByPage reads all Repository
func (dao *RepositoryDAO) GetAllByPage(page int) ([]model.Repository, error) {
	rows, err := dao.db.
		Model(&model.Repository{}).
		Offset(page * dao.pageSize).
		Limit(dao.pageSize).
		Rows()
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return dao.scan(rows)
}

func (dao *RepositoryDAO) scan(rows *sql.Rows) ([]model.Repository, error) {
	repos := make([]model.Repository, 0, dao.pageSize)
	for rows.Next() {
		var repo model.Repository
		err := dao.db.ScanRows(rows, &repo)
		if err != nil {
			return nil, err
		}

		repos = append(repos, repo)
	}
	return repos, nil
}
