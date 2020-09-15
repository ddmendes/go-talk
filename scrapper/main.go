package main

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type repository struct {
	gorm.Model
	ID          int
	Owner       string
	Name        string
	Stars       int
	Forks       int
	Subscribers int
	Watchers    int
	OpenIssues  int
	Size        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var tokenFileName = ".ghtoken"
var pgDsn = "host=localhost port=5432 user=root password=toor dbname=ghdata"

func main() {
	db, err := gorm.Open(postgres.Open(pgDsn), &gorm.Config{})
	db.AutoMigrate(&repository{})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	client := getGitHubClient(ctx)
	scrapRepos(ctx, client, db)
}

func getGitHubClient(ctx context.Context) *github.Client {
	byteToken, err := ioutil.ReadFile(tokenFileName)
	if err != nil {
		panic("Could not read GitHub token")
	}

	sToken := strings.TrimSpace(string(byteToken))
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: sToken},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSource)

	return github.NewClient(tokenClient)
}

func scrapRepos(ctx context.Context, client *github.Client, db *gorm.DB) {
	repo, _, err := client.Repositories.Get(ctx, "ddmendes", "go-talk")
	if err != nil {
		panic(err)
	}
	r := repository{
		Owner:       *repo.Owner.Login,
		Name:        *repo.Name,
		Stars:       *repo.StargazersCount,
		Forks:       *repo.ForksCount,
		Subscribers: *repo.SubscribersCount,
		Watchers:    *repo.WatchersCount,
		OpenIssues:  *repo.OpenIssuesCount,
		Size:        *repo.Size,
	}
	db.Create(&r)
}
