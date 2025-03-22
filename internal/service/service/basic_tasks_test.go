package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Dyleme/Notifier/internal/domain"
)

func TestService_CreateBasicTask(t *testing.T) {
	type fields struct {
		repos       repositories
		notifierJob NotifierJob
		tr          TxManager
	}
	type args struct {
		ctx  context.Context
		task domain.BasicTask
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.BasicTask
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repos:       tt.fields.repos,
				notifierJob: tt.fields.notifierJob,
				tr:          tt.fields.tr,
			}
			got, err := s.CreateBasicTask(tt.args.ctx, tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateBasicTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.CreateBasicTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetBasicTask(t *testing.T) {
	type fields struct {
		repos       repositories
		notifierJob NotifierJob
		tr          TxManager
	}
	type args struct {
		ctx    context.Context
		userID int
		taskID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.BasicTask
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repos:       tt.fields.repos,
				notifierJob: tt.fields.notifierJob,
				tr:          tt.fields.tr,
			}
			got, err := s.GetBasicTask(tt.args.ctx, tt.args.userID, tt.args.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetBasicTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.GetBasicTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ListBasicTasks(t *testing.T) {
	type fields struct {
		repos       repositories
		notifierJob NotifierJob
		tr          TxManager
	}
	type args struct {
		ctx    context.Context
		userID int
		params ListFilterParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []domain.BasicTask
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repos:       tt.fields.repos,
				notifierJob: tt.fields.notifierJob,
				tr:          tt.fields.tr,
			}
			got, err := s.ListBasicTasks(tt.args.ctx, tt.args.userID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ListBasicTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.ListBasicTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_UpdateBasicTask(t *testing.T) {
	type fields struct {
		repos       repositories
		notifierJob NotifierJob
		tr          TxManager
	}
	type args struct {
		ctx    context.Context
		params domain.BasicTask
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.BasicTask
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repos:       tt.fields.repos,
				notifierJob: tt.fields.notifierJob,
				tr:          tt.fields.tr,
			}
			got, err := s.UpdateBasicTask(tt.args.ctx, tt.args.params, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.UpdateBasicTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.UpdateBasicTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_DeleteBasicTask(t *testing.T) {
	type fields struct {
		repos       repositories
		notifierJob NotifierJob
		tr          TxManager
	}
	type args struct {
		ctx    context.Context
		userID int
		taskID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repos:       tt.fields.repos,
				notifierJob: tt.fields.notifierJob,
				tr:          tt.fields.tr,
			}
			if err := s.DeleteBasicTask(tt.args.ctx, tt.args.userID, tt.args.taskID); (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteBasicTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
