package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ErrNotFound = errors.New("link not found")

// Link описывает одну короткую ссылку.
type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Clicks    int       `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

// Store потокобезопасное хранилище ссылок с сохранением в JSON-файл.
type Store struct {
	mu    sync.RWMutex
	links map[string]*Link
	path  string
}

// NewStore загружает данные из файла (если он есть). Пустой path - только память.
func NewStore(path string) (*Store, error) {
	s := &Store{links: make(map[string]*Link), path: path}
	if path == "" {
		return s, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.links); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Create(url string) (*Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var code string
	for {
		c, err := newCode(6)
		if err != nil {
			return nil, err
		}
		if _, taken := s.links[c]; !taken {
			code = c
			break
		}
	}
	l := &Link{Code: code, URL: url, CreatedAt: time.Now().UTC()}
	s.links[code] = l
	return copyLink(l), s.save()
}

// Get возвращает ссылку без изменения счетчика.
func (s *Store) Get(code string) (*Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.links[code]
	if !ok {
		return nil, ErrNotFound
	}
	return copyLink(l), nil
}

// Hit увеличивает счетчик переходов и возвращает ссылку.
func (s *Store) Hit(code string) (*Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.links[code]
	if !ok {
		return nil, ErrNotFound
	}
	l.Clicks++
	return copyLink(l), s.save()
}

// save вызывается под блокировкой. Пишем во временный файл и переименовываем.
func (s *Store) save() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.links, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func copyLink(l *Link) *Link {
	c := *l
	return &c
}

func newCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}
