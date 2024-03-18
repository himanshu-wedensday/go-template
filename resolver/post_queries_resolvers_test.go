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

func TestPosts(
	t *testing.T,
) {
	cases := []struct {
		name       string
		pagination *fm.PostsPagination
		queryInput *fm.PostQueryInput
		wantResp   []*models.Post
		wantErr    bool
	}{
		{
			name:    ErrorFindingPost,
			wantErr: true,
		},
		{
			name:    "pagination",
			wantErr: false,
			pagination: &fm.PostsPagination{
				Limit: 1,
				Page:  1,
			},
			queryInput: &fm.PostQueryInput{
				AuthorID: testutls.MockPost().AuthorID,
			},
			wantResp: testutls.MockPosts(),
		},
		{
			name:     SuccessCase,
			wantErr:  false,
			wantResp: testutls.MockPosts(),
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
					mock.ExpectQuery(regexp.QuoteMeta(`select * from "posts" where "id"=$1`)).
						WithArgs().
						WillReturnError(fmt.Errorf(""))
				}

				if tt.name == "pagination" {
					rows := sqlmock.
						NewRows([]string{"author_id", "post"}).
						AddRow(testutls.MockID, "Hello post")
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT "posts".* FROM "posts" where author_id= 1 LIMIT 1 OFFSET 1;`)).WithArgs().WillReturnRows(rows)

					rowCount := sqlmock.NewRows([]string{"count"}).
						AddRow(1)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "posts" where author_id=1 LIMIT 1;`)).
						WithArgs().
						WillReturnRows(rowCount)

				} else {
					rows := sqlmock.
						NewRows([]string{"id", "author_id", "post"}).
						AddRow(testutls.MockID, testutls.MockAuthor().ID, "Hello Post")
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT "posts".* FROM "posts";`)).WithArgs().WillReturnRows(rows)

					rowCount := sqlmock.NewRows([]string{"count"}).
						AddRow(1)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "posts";`)).
						WithArgs().
						WillReturnRows(rowCount)

				}
				// Define a mock result set for user queries.

				// Define a mock result set.

				// Create a new context with a mock user.
				c := context.Background()
				ctx := context.WithValue(c, testutls.PostKey, testutls.MockPost())

				// Query for users using the resolver and get the response and error.
				response, err := resolver1.Query().
					Posts(ctx, tt.queryInput, tt.pagination)

				// Check if the response matches the expected response length.
				if tt.wantResp != nil &&
					response != nil {
					assert.Equal(t, len(tt.wantResp), len(response.Posts))

				}
				// Check if the error matches the expected error value.
				assert.Equal(t, tt.wantErr, err != nil)
			},
		)
	}
}
