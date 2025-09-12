package http

type Resource[T any] struct {
	Data T `json:"data"`
}

type ResourceMetadata struct {
	Total int64 `json:"total"`
	Limit int64 `json:"limit"`
	Page  int64 `json:"page"`
}

type ResourceCollection[T any] struct {
	Data []T `json:"data"`
}

func NewResource[T any](data T) Resource[T] {
	return Resource[T]{Data: data}
}

func NewResourceCollection[T any](data []T) ResourceCollection[T] {
	return ResourceCollection[T]{Data: data}
}
