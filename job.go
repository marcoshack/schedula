package schedula

// Job ...
type Job struct {
	ID          string            `json:"id"`
	BusinessKey string            `json:"businessKey"`
	CallbackURL string            `json:"callbackURL"`
	Data        map[string]string `json:"data"`
	Timeout     JobTimeout        `json:"timeout"`
}

// JobTimeout ...
type JobTimeout struct {
	Format string `json:"format"`
	Value  string `json:"value"`
}
