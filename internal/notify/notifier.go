package notify

type Notification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Notifier interface {
	Send(n Notification) error
}
