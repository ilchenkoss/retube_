package domain

type Subscriptions struct {
	ID          int
	Subscriber  *User
	SubscribeTo *User
}
