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

// RepositoryService access point to service logics
type RepositoryService struct {
	RepoDAO  *dao.RepositoryDAO
	GHClient *github.Client
	Ctx      context.Context
}

type message struct {
	localRepo  *model.Repository
	remoteRepo *github.Repository
}

// UpdateAll updates all entries
func (s *RepositoryService) UpdateAll() {
	readLocalChan := make(chan []model.Repository)
	readRemoteChan := make(chan *message)
	filterChan := make(chan *message)
	writeChan := make(chan *model.Repository)
	wg := &sync.WaitGroup{}

	wg.Add(5)
	go s.readLocal(readLocalChan, wg)
	go s.fanOut(readLocalChan, readRemoteChan, wg)

	remoteWG := &sync.WaitGroup{}
	go func() {
		remoteWG.Wait()
		close(filterChan)
		wg.Done()
	}()
	for i := 0; i < runtime.NumCPU(); i++ {
		remoteWG.Add(1)
		go s.readRemote(readRemoteChan, filterChan, remoteWG)
	}

	go s.filter(filterChan, writeChan, wg)
	go s.write(writeChan, wg)
	wg.Wait()
}

func (s *RepositoryService) readLocal(out chan<- []model.Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("readLocal finished") }()

	log.Println("readLocal started")

	page := 0
	counter := 0
	keepReading := true
	for keepReading {
		repos, err := s.RepoDAO.GetAllByPage(page)
		switch {
		case err != nil:
			log.Printf("Failed to read local repo: %s", err.Error())
			fallthrough
		case len(repos) <= 0:
			keepReading = false
		default:
			counter += len(repos)
			log.Printf("Read %d repos. Total: %d", len(repos), counter)
			out <- repos
		}
		page++
	}
}

func (s *RepositoryService) fanOut(in <-chan []model.Repository, out chan<- *message, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("fanOut finished") }()

	log.Println("fanOut started")
	for batch := range in {
		for _, repo := range batch {
			msg := &message{
				localRepo: &repo,
			}
			out <- msg
		}
	}
}

func (s *RepositoryService) readRemote(in <-chan *message, out chan<- *message, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { log.Println("readRemote finished") }()

	log.Println("readRemote started")
	for m := range in {
		remoteRepo, _, err := s.GHClient.Repositories.Get(
			s.Ctx, m.localRepo.Owner, m.localRepo.Name)
		if err != nil {
			log.Printf("Could not read %s/%s: %s",
				m.localRepo.Owner, m.localRepo.Name, err.Error())
			continue
		}
		m.remoteRepo = remoteRepo
		out <- m
	}
}

func (s *RepositoryService) filter(in <-chan *message, out chan<- *model.Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	defer func() { log.Println("filter finished") }()

	log.Println("filter started")
	for m := range in {
		if !areReposEquals(m.localRepo, m.remoteRepo) {
			updateRepo(m.localRepo, m.remoteRepo)
			out <- m.localRepo
		}
	}

}

func (s *RepositoryService) write(in <-chan *model.Repository, wg *sync.WaitGroup) {
	counter := 0
	defer wg.Done()
	defer func() { log.Println("write finished") }()

	log.Println("write started")
	for repo := range in {
		s.RepoDAO.Save(repo)
		counter++
		if counter%50 == 0 {
			log.Printf("Wrote %d repositories", counter)
		}
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
