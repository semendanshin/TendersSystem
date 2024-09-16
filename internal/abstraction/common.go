package abstraction

type PaginationOptions struct {
	Limit  int
	Offset int
}

type PaginationOptFunc func(*PaginationOptions) error

func WithLimit(limit int) PaginationOptFunc {
	return func(o *PaginationOptions) error {
		o.Limit = limit
		return nil
	}
}

func WithOffset(offset int) PaginationOptFunc {
	return func(o *PaginationOptions) error {
		o.Offset = offset
		return nil
	}
}

func NewPaginationOptions(options ...PaginationOptFunc) (*PaginationOptions, error) {
	opts := &PaginationOptions{
		Limit:  10,
		Offset: 0,
	}
	for _, opt := range options {
		if err := opt(opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}
