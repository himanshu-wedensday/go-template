package resolver_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	fm "go-template/gqlmodels"
	"go-template/models"
	"go-template/resolver"
	"go-template/testutls"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/stretchr/testify/assert"
)

func TestAuthors(
	t *testing.T,
) {
	cases := []struct {
		name       string
		pagination *fm.AuthorPagination
		wantResp   []*models.Author
		wantErr    bool
	}{
		{
			name:    ErrorFindingAuthor,
			wantErr: true,
		},
		{
			name:    "pagination",
			wantErr: false,
			pagination: &fm.AuthorPagination{
				Limit: 1,
				Page:  1,
			},
			wantResp: testutls.MockAuthors(),
		},
		{
			name:     SuccessCase,
			wantErr:  false,
			wantResp: testutls.MockAuthors(),
		},
	}

	// Create a new instance of the resolver.
	resolver1 := resolver.Resolver{}
	for _, tt := range cases {
		t.Run(
			tt.name,
			func(t *testing.T) {

				// Load environment variables from the .env.local file.
				err := godotenv.Load(
					"../.env.local",
				)
				if err != nil {
					fmt.Print("error loading .env file")
				}

				// Create a new mock database connection.
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

				//fail on finding user case
				if tt.name == ErrorFindingUser {
					mock.ExpectQuery(regexp.QuoteMeta(`select * from "authors" where "id"=$1`)).
						WithArgs().
						WillReturnError(fmt.Errorf(""))
				}

				if tt.name == "pagination" {
					rows := sqlmock.
						NewRows([]string{"id", "first_name", "last_name", "email"}).
						AddRow(testutls.MockID, "First", "Last", testutls.MockEmail)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT "authors".* FROM "authors" LIMIT 1 OFFSET 1;`)).WithArgs().WillReturnRows(rows)

					rowCount := sqlmock.NewRows([]string{"count"}).
						AddRow(1)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "authors" LIMIT 1;`)).
						WithArgs().
						WillReturnRows(rowCount)

				} else {
					rows := sqlmock.
						NewRows([]string{"id", "email", "first_name", "last_name"}).
						AddRow(testutls.MockID, testutls.MockEmail, "First", "Last")
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT "authors".* FROM "authors";`)).WithArgs().WillReturnRows(rows)

					rowCount := sqlmock.NewRows([]string{"count"}).
						AddRow(1)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "authors";`)).
						WithArgs().
						WillReturnRows(rowCount)

				}
				// Define a mock result set for user queries.

				// Define a mock result set.

				// Create a new context with a mock user.
				c := context.Background()
				ctx := context.WithValue(c, testutls.AuthorKey, testutls.MockAuthor())

				// Query for users using the resolver and get the response and error.
				response, err := resolver1.Query().
					Authors(ctx, tt.pagination)

				// Check if the response matches the expected response length.
				if tt.wantResp != nil &&
					response != nil {
					assert.Equal(t, len(tt.wantResp), len(response.Authors))

				}
				// Check if the error matches the expected error value.
				assert.Equal(t, tt.wantErr, err != nil)
			},
		)
	}
}
