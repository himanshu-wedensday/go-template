package daos_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"go-template/daos"
	"go-template/internal/config"
	"go-template/models"
	"go-template/pkg/utl/convert"
	"go-template/testutls"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const ErrorFindingAuthor = "Fail on finding Author"

func TestCreateAuthorTx(t *testing.T) {

	cases := []struct {
		name string
		req  models.Author
		err  error
	}{
		{
			name: "Passing author type value",
			req: models.Author{
				ID:        testutls.MockAuthor().ID,
				Email:     testutls.MockAuthor().Email,
				FirstName: testutls.MockAuthor().FirstName,
				LastName:  testutls.MockAuthor().LastName,
			},
			err: nil,
		},
	}

	for _, tt := range cases {
		// Inject mock instance into boil.

		err := config.LoadEnvWithFilePrefix(convert.StringToPointerString("./../"))
		if err != nil {
			log.Fatal(err)
		}
		mock, db, _ := testutls.SetupMockDB(t)
		oldDB := boil.GetDB()
		defer func() {
			db.Close()
			boil.SetDB(oldDB)
		}()
		boil.SetDB(db)

		rows := sqlmock.NewRows([]string{
			"first_name",
			"last_name",
			"email",
			"deleted_at",
		}).AddRow(
			testutls.MockAuthor().FirstName,
			testutls.MockAuthor().LastName,
			testutls.MockAuthor().Email,
			testutls.MockAuthor().DeletedAt,
		)
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "authors"`)).
			WithArgs().
			WillReturnRows(rows)

		t.Run(tt.name, func(t *testing.T) {
			_, err := daos.CreateAuthor(tt.req, context.Background())
			if err != nil {
				assert.Equal(t, true, tt.err != nil)
			} else {
				assert.Equal(t, err, tt.err)
			}
		})
	}
}

func TestFindAuthorByID(t *testing.T) {

	cases := []struct {
		name string
		req  int
		err  error
	}{
		{
			name: "Passing a author ID",
			req:  1,
			err:  nil,
		},
	}

	for _, tt := range cases {
		err := config.LoadEnv()
		if err != nil {
			fmt.Print("error loading .env file")
		}

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		// Inject mock instance into boil.
		oldDB := boil.GetDB()
		defer func() {
			db.Close()
			boil.SetDB(oldDB)
		}()
		boil.SetDB(db)

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

		mock.ExpectQuery(regexp.QuoteMeta(`select * from "authors" where "id"=$1`)).
			WithArgs().
			WillReturnRows(rows)

		t.Run(tt.name, func(t *testing.T) {
			_, err := daos.FindAuthorWithId(strconv.Itoa(tt.req), context.Background())
			assert.Equal(t, err, tt.err)

		})
	}
}

func TestUpdateAuthorTx(t *testing.T) {

	cases := []struct {
		name string
		req  models.Author
		err  error
	}{
		{
			name: "Passing author type value",
			req:  models.Author{},
			err:  nil,
		},
	}

	for _, tt := range cases {
		err := config.LoadEnv()
		if err != nil {
			fmt.Print("error loading .env file")
		}

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		// Inject mock instance into boil.
		oldDB := boil.GetDB()
		defer func() {
			db.Close()
			boil.SetDB(oldDB)
		}()
		boil.SetDB(db)

		result := driver.Result(driver.RowsAffected(1))
		// get access_token
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "authors" `)).
			WillReturnResult(result)

		t.Run(tt.name, func(t *testing.T) {
			_, err := daos.UpdateAuthor(tt.req, context.Background())
			assert.Equal(t, err, tt.err)
		})
	}
}

func TestDeleteAuthor(t *testing.T) {

	cases := []struct {
		name string
		req  models.Author
		err  error
	}{
		{
			name: "Passing author type value",
			req:  models.Author{},
			err:  nil,
		},
	}

	for _, tt := range cases {
		err := config.LoadEnv()
		if err != nil {
			fmt.Print("error loading .env file")
		}

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		// Inject mock instance into boil.
		oldDB := boil.GetDB()
		defer func() {
			db.Close()
			boil.SetDB(oldDB)
		}()
		boil.SetDB(db)

		// delete user
		result := driver.Result(driver.RowsAffected(1))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "authors" WHERE "id"=$1`)).
			WillReturnResult(result)

		t.Run(tt.name, func(t *testing.T) {
			_, err := daos.DeleteAuthor(tt.req, context.Background())
			assert.Equal(t, err, tt.err)
		})
	}
}

func TestFindAllAuthorsWithCount(t *testing.T) {

	oldDB := boil.GetDB()
	err := config.LoadEnvWithFilePrefix(convert.StringToPointerString("./../"))
	if err != nil {
		log.Fatal(err)
	}
	mock, db, _ := testutls.SetupMockDB(t)

	cases := []struct {
		name      string
		err       error
		dbQueries []testutls.QueryData
	}{
		{
			name: "Failed to find all authors with count",
			err:  fmt.Errorf("sql: no rows in sql"),
		},
		{
			name: "Successfully find all authors with count",
			err:  nil,
			dbQueries: []testutls.QueryData{
				{
					Query: `SELECT "authors".* FROM "authors";`,
					DbResponse: sqlmock.NewRows([]string{"id", "email", "first_name", "last_name"}).AddRow(
						testutls.MockID,
						testutls.MockEmail,
						testutls.MockAuthor().FirstName,
						testutls.MockAuthor().LastName),
				},
				{
					Query:      `SELECT COUNT(*) FROM "authors";`,
					DbResponse: sqlmock.NewRows([]string{"count"}).AddRow(testutls.MockCount),
				},
			},
		},
	}

	for _, tt := range cases {

		if tt.err != nil {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT "authors".* FROM "authors";`)).
				WithArgs().
				WillReturnError(fmt.Errorf("this is some error"))
		}

		for _, dbQuery := range tt.dbQueries {
			mock.ExpectQuery(regexp.QuoteMeta(dbQuery.Query)).
				WithArgs().
				WillReturnRows(dbQuery.DbResponse)
		}

		t.Run(tt.name, func(t *testing.T) {
			res, c, err := daos.FindAllAuthorsWithCount([]qm.QueryMod{}, context.Background())
			if err != nil {
				assert.Equal(t, true, tt.err != nil)
			} else {
				assert.Equal(t, err, tt.err)
				assert.Equal(t, testutls.MockCount, c)
				assert.Equal(t, res[0].Email, null.StringFrom(testutls.MockEmail))
				assert.Equal(t, res[0].ID, int(testutls.MockID))
			}
		})
	}
	boil.SetDB(oldDB)
	db.Close()
}
