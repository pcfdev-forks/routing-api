package db

import "github.com/jinzhu/gorm"

//go:generate counterfeiter -o fakes/fake_client.go . Client
type Client interface {
	Close() error
	Where(query interface{}, args ...interface{}) Client
	Create(value interface{}) (int64, error)
	Delete(value interface{}, where ...interface{}) (int64, error)
	Save(value interface{}) (int64, error)
	Update(attrs ...interface{}) (int64, error)
	First(out interface{}, where ...interface{}) error
	Find(out interface{}, where ...interface{}) error
	AutoMigrate(values ...interface{}) error
	Begin() Client
	Rollback() error
	Commit() error
	HasTable(value interface{}) bool
}

type gormClient struct {
	db *gorm.DB
}

func NewGormClient(db *gorm.DB) Client {
	return &gormClient{db: db}
}

func (c *gormClient) Close() error {
	return c.db.Close()
}

func (c *gormClient) Where(query interface{}, args ...interface{}) Client {
	var newClient gormClient
	newClient.db = c.db.Where(query, args...)
	return &newClient
}

func (c *gormClient) Create(value interface{}) (int64, error) {
	newDb := c.db.Create(value)
	return newDb.RowsAffected, newDb.Error
}

func (c *gormClient) Delete(value interface{}, where ...interface{}) (int64, error) {
	newDb := c.db.Delete(value, where...)
	return newDb.RowsAffected, newDb.Error
}

func (c *gormClient) Save(value interface{}) (int64, error) {
	newDb := c.db.Save(value)
	return newDb.RowsAffected, newDb.Error
}

func (c *gormClient) Update(attrs ...interface{}) (int64, error) {
	newDb := c.db.Update(attrs...)
	return newDb.RowsAffected, newDb.Error
}

func (c *gormClient) First(out interface{}, where ...interface{}) error {
	return c.db.First(out, where...).Error
}

func (c *gormClient) Find(out interface{}, where ...interface{}) error {
	return c.db.Find(out, where...).Error
}

func (c *gormClient) AutoMigrate(values ...interface{}) error {
	return c.db.AutoMigrate(values...).Error
}

func (c *gormClient) Begin() Client {
	var newClient gormClient
	newClient.db = c.db.Begin()
	return &newClient
}

func (c *gormClient) Rollback() error {
	return c.db.Rollback().Error
}

func (c *gormClient) Commit() error {
	return c.db.Commit().Error
}

func (c *gormClient) HasTable(value interface{}) bool {
	return c.db.HasTable(value)
}
