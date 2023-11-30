package schemagen

import (
	"github.com/go-pg/pg/v10/orm"
	"github.com/monetrapp/rest-api/pkg/internal/testutils"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTableQuery_String(t *testing.T) {
	t.Run("with unique", func(t *testing.T) {
		log := testutils.GetLog(t)

		type Login struct {
			tableName string `pg:"logins"`

			LoginId      uint64 `json:"loginId" pg:"login_id,notnull,pk,type:'bigserial'"`
			Email        string `json:"email" pg:"email,notnull,unique"`
			PasswordHash string `json:"-" pg:"password_hash,notnull"`
		}

		createTable := NewCreateTableQuery(&Login{}, orm.CreateTableOptions{
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})

		query := createTable.String()
		assert.NotEmpty(t, query, "query should not be empty")

		log.Debug(query)
	})

	t.Run("backwards fk", func(t *testing.T) {
		log := testutils.GetLog(t)

		createTable := NewCreateTableQuery(&models.User{}, orm.CreateTableOptions{
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})

		query := createTable.String()
		assert.NotEmpty(t, query, "query should not be empty")

		log.Debug(query)
	})
}
