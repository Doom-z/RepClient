package run

type HasDomainID interface {
	GetDomainID() string
}

type SaveTask struct {
	Data   any
	Path   string
	Format string
}
