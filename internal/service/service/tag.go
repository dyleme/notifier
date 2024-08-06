package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type TagsRepository interface {
	Add(ctx context.Context, tag string, userID int) (domains.Tag, error)
	Get(ctx context.Context, tagID int) (domains.Tag, error)
	List(ctx context.Context, userID int, listParams ListParams) ([]domains.Tag, error)
	Update(ctx context.Context, tagID int, name string) error
	Delete(ctx context.Context, tagID int) error
}

func (s *Service) AddTag(ctx context.Context, tag string, userID int) (domains.Tag, error) {
	var createdTag domains.Tag
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		createdTag, err = s.repos.tags.Add(ctx, tag, userID)
		if err != nil {
			return fmt.Errorf("add tag: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domains.Tag{}, err
	}

	return createdTag, nil
}

func (s *Service) GetTag(ctx context.Context, tagID, userID int) (domains.Tag, error) {
	var tag domains.Tag
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		tag, err = s.repos.tags.Get(ctx, tagID)
		if err != nil {
			return fmt.Errorf("get tag: %w", err)
		}

		if tag.BelongsTo(userID) != nil {
			return fmt.Errorf("get tag: %w", serverrors.NewBusinessLogicError("tag does not belong to user"))
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domains.Tag{}, err
	}

	return tag, nil
}

func (s *Service) ListTags(ctx context.Context, userID int, listParams ListParams) ([]domains.Tag, error) {
	var tags []domains.Tag
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		tags, err = s.repos.tags.List(ctx, userID, listParams)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return nil, err
	}

	return tags, nil
}

func (s *Service) DeleteTag(ctx context.Context, tagID, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		tag, err := s.repos.tags.Get(ctx, tagID)
		if err != nil {
			return fmt.Errorf("get tag: %w", err)
		}

		if tag.BelongsTo(userID) != nil {
			return fmt.Errorf("delete tag: %w", serverrors.NewBusinessLogicError("tag does not belong to user"))
		}

		err = s.repos.tags.Delete(ctx, tagID)
		if err != nil {
			return fmt.Errorf("delete tag: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return err
	}

	return nil
}

func (s *Service) UpdateTag(ctx context.Context, tagID int, name string, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		tag, err := s.repos.tags.Get(ctx, tagID)
		if err != nil {
			return fmt.Errorf("get tag: %w", err)
		}

		if tag.BelongsTo(userID) != nil {
			return fmt.Errorf("update tag: %w", serverrors.NewBusinessLogicError("tag does not belong to user"))
		}

		err = s.repos.tags.Update(ctx, tagID, name)
		if err != nil {
			return fmt.Errorf("update tag: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return err
	}

	return nil
}
