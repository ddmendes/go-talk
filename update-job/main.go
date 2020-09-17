package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/ddmendes/go-talk/update-job/dao"
	"github.com/ddmendes/go-talk/update-job/data"
	"github.com/ddmendes/go-talk/update-job/model"
	"github.com/ddmendes/go-talk/update-job/service"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func main() {
	initConfig()
	initDB()

	ctx := context.Background()
	ghClient := newGitHubClient(ctx)
	repoDAO, err := dao.NewRepositoryDAO()
	if err != nil {
		log.Fatal(err)
	}
	repoService := &service.RepositoryService{
		RepoDAO:  repoDAO,
		GHClient: ghClient,
		Ctx:      ctx,
	}
	repoService.UpdateAll()
}

func initConfig() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Could not open config file config.json: %s", err.Error())
	}
	defer configFile.Close()
	configBuffer := bufio.NewReader(configFile)
	err = data.LoadConfiguration(configBuffer)
	if err != nil {
		log.Fatalf("Failed to read configuration file: %s", err.Error())
	}
}

func initDB() {
	data.SetDSN(data.Config.PostgresDSN)
	db, err := data.GetDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err.Error())
	}
	db.AutoMigrate(&model.Repository{})
}

func newGitHubClient(ctx context.Context) *github.Client {
	sToken := strings.TrimSpace(string(data.Config.GitHubToken))
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: sToken},
	)
	tokenClient := oauth2.NewClient(ctx, tokenSource)

	return github.NewClient(tokenClient)
}
