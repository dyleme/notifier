package service

type ListParams struct {
	Offset int
	Limit  int
}

type ListFilterParams struct {
	ListParams ListParams
	TagIDs     []int
}
