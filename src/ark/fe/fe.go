package fe

import "ark/store"

// Service ...
type Service interface {
	Update([]*store.Route) error
}
