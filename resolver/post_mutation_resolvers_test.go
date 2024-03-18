package resolver_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"go-template/daos"
	fm "go-template/gqlmodels"
	"go-template/internal/config"
	"go-template/models"
	"go-template/pkg/utl/convert"
	"go-template/pkg/utl/throttle"
	"go-template/resolver"
	"go-template/testutls"
	"log"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/joho/godotenv"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	ErrorFromCreatePost = "Error Creating Post"
	ErrorFindingPost    = "Error Finding Post"
	ErrorDeletePost     = "Error Delete Post"
	ErrorUpdatePost     = "Error Update Post"
)

func TestCreatePost(
	t *testing.T,
) {
	cases := []struct {
		name     string
		req      fm.PostCreateInput
		wantResp *fm.Post
		wantErr  bool
	}{
		{
			name:    ErrorFromCreatePost,
			req:     fm.PostCreateInput{},
			wantErr: true,
		},
		{
			name:    ErrorFromThrottleCheck,
			req:     fm.PostCreateInput{},
			wantErr: true,
		},
		{
			name:    ErrorFromConfig,
			req:     fm.PostCreateInput{},
			wantErr: true,
		},
		{
			name: SuccessCase,
			req: fm.PostCreateInput{
				Post:     testutls.MockPost().Post,
				AuthorID: strconv.Itoa(testutls.MockPost().AuthorID),
			},
			wantResp: &fm.Post{
				ID:        fmt.Sprint(testutls.MockPost().ID),
				Post:      convert.StringToPointerString(testutls.MockPost().Post),
				Author:    &fm.Author{ID: strconv.Itoa(testutls.MockPost().AuthorID)},
				DeletedAt: convert.NullDotTimeToPointerInt(testutls.MockAuthor().DeletedAt),
				UpdatedAt: convert.NullDotTimeToPointerInt(testutls.MockAuthor().UpdatedAt),
			},
			wantErr: false,
		},
	}

	resolver1 := resolver.Resolver{}
	for _, tt := range cases {
		t.Run(
			tt.name,
			func(t *testing.T) {
				if tt.name == ErrorFromThrottleCheck {
					patch := gomonkey.ApplyFunc(throttle.Check, func(ctx context.Context, limit int, dur time.Duration) error {
						return fmt.Errorf("Internal error")
					})
					defer patch.Reset()
				}

				if tt.name == ErrorFromConfig {
					patch := gomonkey.ApplyFunc(config.Load, func() (*config.Configuration, error) {
						return nil, fmt.Errorf("error in loading config")
					})
					defer patch.Reset()

				}

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

				if tt.name == ErrorFromCreatePost {
					// insert new Author
					mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "posts"`)).
						WithArgs().
						WillReturnError(fmt.Errorf(""))
				}
				// insert new Author
				rows := sqlmock.NewRows([]string{
					"id",
				}).
					AddRow(
						testutls.MockPost().ID,
					)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "posts"`)).
					WithArgs(
						testutls.MockPost().Post,
						testutls.MockPost().AuthorID,
						"",
						AnyTime{},
						AnyTime{},
					).
					WillReturnRows(rows)

				c := context.Background()
				response, err := resolver1.Mutation().
					CreatePost(c, tt.req)
				if tt.wantResp != nil {
					assert.Equal(t, tt.wantResp, response)
				}
				assert.Equal(t, tt.wantErr, err != nil)
			},
		)
	}
}

func TestUpdatePost(
	t *testing.T,
) {
	cases := []struct {
		name     string
		req      *fm.PostUpdateInput
		wantResp *fm.Post
		wantErr  bool
	}{
		{
			name:    ErrorFindingPost,
			req:     &fm.PostUpdateInput{},
			wantErr: true,
		},
		{
			name: ErrorUpdatePost,
			req: &fm.PostUpdateInput{
				Post: &testutls.MockPost().Post,
			},
			wantErr: true,
		},
		{
			name: SuccessCase,
			req: &fm.PostUpdateInput{
				Post: &testutls.MockPost().Post,
			},
			wantResp: &fm.Post{
				ID:   "0",
				Post: &testutls.MockPost().Post,
			},
			wantErr: false,
		},
	}

	resolver1 := resolver.Resolver{}
	for _, tt := range cases {
		t.Run(
			tt.name,
			func(t *testing.T) {

				if tt.name == ErrorUpdatePost {

					patch := gomonkey.ApplyFunc(daos.UpdatePost,
						func(post models.Post, ctx context.Context) (models.Post, error) {
							return post, fmt.Errorf("error for update Author")
						})
					defer patch.Reset()
				}
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

				if tt.name == ErrorFindingPost {
					mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "posts"`)).WithArgs().WillReturnError(fmt.Errorf(""))
				}

				rows := sqlmock.NewRows([]string{"post"}).AddRow(testutls.MockPost().Post)
				mock.ExpectQuery(regexp.QuoteMeta(`select * from "posts"`)).WithArgs(0).WillReturnRows(rows)

				// update Authors with new information
				result := driver.Result(driver.RowsAffected(1))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "posts"`)).WillReturnResult(result)

				c := context.Background()
				ctx := context.WithValue(c, testutls.PostKey, testutls.MockPost())
				response, err := resolver1.Mutation().UpdatePost(ctx, *tt.req)
				if tt.wantResp != nil &&
					response != nil {
					assert.Equal(t, tt.wantResp, response)
				}
				assert.Equal(t, tt.wantErr, err != nil)
			},
		)
	}
}

func TestDeletePost(
	t *testing.T,
) {
	cases := []struct {
		name     string
		wantResp *fm.PostDeletePayload
		wantErr  bool
	}{
		{
			name:    ErrorFindingPost,
			wantErr: true,
		},
		{
			name:    ErrorDeletePost,
			wantErr: true,
		},
		{
			name: SuccessCase,
			wantResp: &fm.PostDeletePayload{
				ID: "0",
			},
			wantErr: false,
		},
	}

	resolver1 := resolver.Resolver{}
	for _, tt := range cases {
		t.Run(
			tt.name,
			func(t *testing.T) {
				if tt.name == ErrorDeletePost {

					patch := gomonkey.ApplyFunc(daos.DeletePost,
						func(post models.Post, ctx context.Context) (int64, error) {
							return 0, fmt.Errorf("error for delete post")
						})
					defer patch.Reset()
				}

				err := godotenv.Load(
					"../.env.local",
				)
				if err != nil {
					fmt.Print("error loading .env file")
				}
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				oldDB := boil.GetDB()
				defer func() {
					db.Close()
					boil.SetDB(oldDB)
				}()
				boil.SetDB(db)

				if tt.name == ErrorFindingAuthor {
					mock.ExpectQuery(regexp.QuoteMeta(`select * from "posts" where "id"=$1`)).
						WithArgs().
						WillReturnError(fmt.Errorf(""))
				}
				// get Author by id
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery(regexp.QuoteMeta(`select * from "posts" where "id"=$1`)).
					WithArgs().
					WillReturnRows(rows)
				// delete Author
				result := driver.Result(driver.RowsAffected(1))
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "posts" WHERE "id"=$1`)).
					WillReturnResult(result)

				c := context.Background()
				ctx := context.WithValue(c, testutls.PostKey, testutls.MockPost())
				response, err := resolver1.Mutation().
					DeletePost(ctx, fm.PostDeleteInput{ID: "1"})
				if tt.wantResp != nil {
					assert.Equal(t, tt.wantResp, response)
				}
				assert.Equal(t, tt.wantErr, err != nil)
			},
		)
	}
}
