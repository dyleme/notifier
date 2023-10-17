package domains

type Task struct {
	ID       int
	UserID   int
	Text     string
	Archived bool
	Periodic bool
}
