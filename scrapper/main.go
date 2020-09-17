package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
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
	log.SetOutput(os.Stdout)
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
	ghReposChan := make(chan []*github.Repository)
	ghRepoChan := make(chan *github.Repository)
	repoChan := make(chan *repository)
	reposChan := make(chan []*repository)
	count := make(chan int)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(6)

	go readRepos(ctx, client, ghReposChan, done, wg)
	go readRepo(ctx, client, ghReposChan, ghRepoChan, wg)
	go convertRepos(ghRepoChan, repoChan, wg)
	go groupRepos(100, repoChan, reposChan, wg)
	go saveRepos(db, reposChan, count, wg)
	go counter(1000000, count, done, wg)
	wg.Wait()
}

func readRepos(ctx context.Context, client *github.Client,
	out chan<- []*github.Repository, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	log.Printf("readRepos started")
	options := &github.RepositoryListAllOptions{
		Since: 0,
	}

	keepGoing := true
	for keepGoing {
		repos, _, err := client.Repositories.ListAll(ctx, options)
		if err != nil || len(repos) == 0 {
			break
		}
		options.Since = *repos[len(repos)-1].ID
		log.Printf("Read %d repos from GitHub", len(repos))
		select {
		case out <- repos:
		case <-done:
			log.Print("readRepos: received done signal")
			keepGoing = false
		}
	}
	log.Printf("readRepos finished")
}

func readRepo(ctx context.Context, client *github.Client,
	in <-chan []*github.Repository, out chan<- *github.Repository,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	log.Printf("readRepo started")
	for ghRepos := range in {
		for _, ghRepo := range ghRepos {
			repo, _, err := client.Repositories.GetByID(ctx, *ghRepo.ID)
			if err != nil {
				continue
			}

			out <- repo
		}
	}
	log.Printf("readRepo finished")
}

func convertRepos(in <-chan *github.Repository, out chan<- *repository,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	log.Printf("convertRepos started")
	for ghRepo := range in {
		repo := &repository{
			Owner:       *ghRepo.Owner.Login,
			Name:        *ghRepo.Name,
			Stars:       *ghRepo.StargazersCount,
			Forks:       *ghRepo.ForksCount,
			Subscribers: *ghRepo.SubscribersCount,
			Watchers:    *ghRepo.WatchersCount,
			OpenIssues:  *ghRepo.OpenIssuesCount,
			Size:        *ghRepo.Size,
		}
		out <- repo
	}
	log.Printf("convertRepos finished")
}

func groupRepos(batchSize int, in <-chan *repository,
	out chan<- []*repository, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	log.Printf("groupRepos started. Batch size: %d entries", batchSize)
	repos := make([]*repository, 0, batchSize)
	for repo := range in {
		repos = append(repos, repo)
		if len(repos) == cap(repos) {
			out <- repos
			repos = make([]*repository, 0, batchSize)
		}
	}

	if len(repos) > 0 {
		out <- repos
	}
	log.Printf("groupRepos finished")
}

func saveRepos(db *gorm.DB, in <-chan []*repository,
	count chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(count)

	log.Printf("saveRepos started")
	for repos := range in {
		db.Create(repos)
		count <- len(repos)
	}
	log.Printf("saveRepos finished")
}

func counter(threshold int, in <-chan int,
	done chan<- struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(done)

	log.Printf("counter started. Thershold: %d", threshold)
	counter := 0
	for count := range in {
		counter += count
		log.Printf("Counter: %d | Threshold: %d\n", counter, threshold)
		if counter >= threshold {
			break
		}
	}
	log.Print("counter finished")
}
