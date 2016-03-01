package models

type Repository struct {
	Name string
	Feed string
}

func (r *Repository) GetFeedHost() (host string) {
	return
}

func (r *Repository) GetFeedRepository() (host string) {
	return
}
