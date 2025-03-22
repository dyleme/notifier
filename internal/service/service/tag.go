package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type TagsRepository interface {
	Add(ctx context.Context, tag string, userID int) (domain.Tag, error)
	Get(ctx context.Context, tagID int) (domain.Tag, error)
	List(ctx context.Context, userID int, listParams ListParams) ([]domain.Tag, error)
	Update(ctx context.Context, tagID int, name string) error
	Delete(ctx context.Context, tagID int) error
}

func (s *Service) AddTag(ctx context.Context, tag string, userID int) (domain.Tag, error) {
	log.Ctx(ctx).Debug("adding tag", "tag", tag, "userID", userID)
	var createdTag domain.Tag
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

		return domain.Tag{}, err
	}

	return createdTag, nil
}

func (s *Service) GetTag(ctx context.Context, tagID, userID int) (domain.Tag, error) {
	log.Ctx(ctx).Debug("getting tag", "tagID", tagID, "userID", userID)
	var tag domain.Tag
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

		return domain.Tag{}, err
	}

	return tag, nil
}

func (s *Service) ListTags(ctx context.Context, userID int, listParams ListParams) ([]domain.Tag, error) {
	log.Ctx(ctx).Debug("list tags", "userID", userID, "listparams", listParams)
	var tags []domain.Tag
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
	log.Ctx(ctx).Debug("delete tag", "userID", userID, "tagID", tagID)
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
	log.Ctx(ctx).Debug("update tags", "tagID", tagID, "name", name, "userID", userID)
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
