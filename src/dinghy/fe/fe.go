package fe

import "dinghy/store"

// Service ...
type Service interface {
	Update([]*store.Route) error
}
