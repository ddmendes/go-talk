package service

import (
	"context"
	"log"
	"runtime"
	"sync"

	"github.com/ddmendes/go-talk/update-job/dao"
	"github.com/ddmendes/go-talk/update-job/model"
	"github.com/google/go-github/v32/github"
)

// RepositoryService access point to service
type RepositoryService struct {
	RepoDAO  *dao.RepositoryDAO
	GHClient *github.Client
	Ctx      context.Context
}

type message struct {
	localRepo  *model.Repository
	remoteRepo *github.Repository
}

// UpdateAll updates all repositories
func (s *RepositoryService) UpdateAll() {
	localReposChan := make(chan []model.Repository)
	remoteReposChan := make(chan *message)
	filterChan := make(chan *message)
	writeChan := make(chan *model.Repository)
	wg := &sync.WaitGroup{}

	wg.Add(4)
	go s.readLocal(localReposChan, wg)
	go s.fanOut(localReposChan, remoteReposChan, wg)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go s.readRemote(remoteReposChan, filterChan, wg)
	}
	go s.filter(filterChan, writeChan, wg)
	go s.write(writeChan, wg)
	wg.Wait()
}

func (s *RepositoryService) readLocal(out chan<- []model.Repository,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("readLocal finished") }()

	log.Println("readLocal started")
	page := 0
	keepReading := true
	counter := 0
	for keepReading {
		repos, err := s.RepoDAO.GetAllByPage(page)
		switch {
		case err != nil:
			log.Printf("Failed to read page %d: %s", page, err.Error())
			fallthrough
		case len(repos) <= 0:
			keepReading = false
		default:
			counter += len(repos)
			log.Printf("readLocal: read %d entries from local. Total: %d",
				len(repos), counter)
			out <- repos
		}
		page++
	}
}

func (s *RepositoryService) fanOut(in <-chan []model.Repository,
	out chan<- *message, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("fanOut finished") }()

	log.Println("fanOut started")
	for repos := range in {
		for _, repo := range repos {
			out <- &message{localRepo: &repo}
		}
	}
}

func (s *RepositoryService) readRemote(in <-chan *message,
	out chan<- *message, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("readRemote finished") }()

	log.Println("readRemote started")
	for message := range in {
		remoteRepo, _, err := s.GhClient.Repositories.Get(
			s.Ctx, message.localRepo.Owner, message.localRepo.Name)
		if err != nil {
			log.Printf("Failed to read repo %s/%s from GitHub: %s",
				message.localRepo.Owner, message.localRepo.Name, err.Error())
			continue
		}
		message.remoteRepo = remoteRepo
		out <- message
	}
}

func (s *RepositoryService) filter(in <-chan *message,
	out chan<- *model.Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("filter finished") }()

	log.Println("filter started")
	for message := range in {
		if !areReposEquals(message.localRepo, message.remoteRepo) {
			updateRepo(message.localRepo, message.remoteRepo)
			out <- message.localRepo
		}
	}
}

func (s *RepositoryService) write(
	in <-chan *model.Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { log.Println("write finished") }()

	log.Println("write started")
	counter := 0
	for repo := range in {
		counter++
		if counter%50 == 0 {
			log.Printf("write: saved %d repos", counter)
		}
		s.RepoDAO.Save(repo)
	}
}

func areReposEquals(l *model.Repository, r *github.Repository) bool {
	return l.Stars == *r.StargazersCount &&
		l.Forks == *r.ForksCount &&
		l.Subscribers == *r.SubscribersCount &&
		l.Watchers == *r.WatchersCount &&
		l.OpenIssues == *r.OpenIssuesCount &&
		l.Size == *r.Size
}

func updateRepo(l *model.Repository, r *github.Repository) {
	l.Stars = *r.StargazersCount
	l.Forks = *r.ForksCount
	l.Subscribers = *r.SubscribersCount
	l.Watchers = *r.WatchersCount
	l.OpenIssues = *r.OpenIssuesCount
	l.Size = *r.Size
}
