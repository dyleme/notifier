package domain

type TaskType string

var PossibleTaskTypes = []TaskType{
	PeriodicTaskType,
	BasicTaskType,
}
