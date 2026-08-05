package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/repo/pg"
)

type PluginRecordUsecase struct {
	records    *pg.PluginRecordRepository
	groups     *pg.PluginRecordGroupRepository
	accessRepo *pg.UserAccessRepository
	logger     *log.Logger
}

func NewPluginRecordUsecase(records *pg.PluginRecordRepository, groups *pg.PluginRecordGroupRepository, accessRepo *pg.UserAccessRepository, logger *log.Logger) *PluginRecordUsecase {
	return &PluginRecordUsecase{records: records, groups: groups, accessRepo: accessRepo, logger: logger.WithModule("usecase.plugin_record")}
}

type PluginRecordWrite struct {
	KBID       string
	PluginID   string
	RecordType string
	Payload    domain.PluginRecordPayload
	Access     domain.PluginRecordAccess
}

type PluginRecordGroupWrite struct {
	KBID          string
	Name          string
	MemberUserIDs []string
}

func (u *PluginRecordUsecase) ListGroups(ctx context.Context, kbID string) ([]domain.PluginRecordGroup, error) {
	userID, _, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	controller, err := u.isKnowledgeBaseController(ctx, kbID, userID)
	if err != nil {
		return nil, err
	}
	return u.groups.ListForUser(ctx, kbID, userID, controller)
}

func (u *PluginRecordUsecase) CreateGroup(ctx context.Context, input PluginRecordGroupWrite) (*domain.PluginRecordGroup, error) {
	userID, _, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	controller, err := u.isKnowledgeBaseController(ctx, input.KBID, userID)
	if err != nil {
		return nil, err
	}
	if !controller {
		return nil, domain.ErrPermissionDenied
	}
	group := &domain.PluginRecordGroup{KBID: strings.TrimSpace(input.KBID), Name: strings.TrimSpace(input.Name), MemberUserIDs: input.MemberUserIDs, CreatedBy: userID}
	if err := u.validateGroupMembers(ctx, group); err != nil {
		return nil, err
	}
	if err := u.groups.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (u *PluginRecordUsecase) UpdateGroup(ctx context.Context, groupID string, input PluginRecordGroupWrite) (*domain.PluginRecordGroup, error) {
	userID, _, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	controller, err := u.isKnowledgeBaseController(ctx, input.KBID, userID)
	if err != nil {
		return nil, err
	}
	if !controller {
		return nil, domain.ErrPermissionDenied
	}
	id, err := strconv.ParseInt(strings.TrimSpace(groupID), 10, 64)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("invalid plugin record group id")
	}
	group := &domain.PluginRecordGroup{ID: id, KBID: strings.TrimSpace(input.KBID), Name: strings.TrimSpace(input.Name), MemberUserIDs: input.MemberUserIDs}
	if err := u.validateGroupMembers(ctx, group); err != nil {
		return nil, err
	}
	if err := u.groups.Update(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (u *PluginRecordUsecase) List(ctx context.Context, kbID, pluginID, recordType string) ([]domain.PluginRecord, error) {
	userID, groupIDs, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	return u.records.ListVisible(ctx, kbID, pluginID, recordType, userID, groupIDs)
}

func (u *PluginRecordUsecase) Create(ctx context.Context, input PluginRecordWrite) (*domain.PluginRecord, error) {
	userID, _, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	record := &domain.PluginRecord{
		ID:          uuid.NewString(),
		KBID:        strings.TrimSpace(input.KBID),
		PluginID:    strings.TrimSpace(input.PluginID),
		RecordType:  strings.TrimSpace(input.RecordType),
		OwnerUserID: userID,
		Payload:     input.Payload,
		Access:      input.Access,
	}
	if err := u.validateWrite(ctx, record); err != nil {
		return nil, err
	}
	if err := u.records.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (u *PluginRecordUsecase) Update(ctx context.Context, recordID string, input PluginRecordWrite) (*domain.PluginRecord, error) {
	userID, groupIDs, err := u.viewer(ctx)
	if err != nil {
		return nil, err
	}
	controller, err := u.isKnowledgeBaseController(ctx, input.KBID, userID)
	if err != nil {
		return nil, err
	}
	record, err := u.records.GetByID(ctx, input.KBID, input.PluginID, input.RecordType, recordID, false)
	if err != nil {
		return nil, err
	}
	if !record.CanEdit(userID, controller, groupIDs) {
		return nil, domain.ErrPermissionDenied
	}
	record.Payload = input.Payload
	record.Access = input.Access
	if err := u.validateWrite(ctx, record); err != nil {
		return nil, err
	}
	if err := u.records.Update(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (u *PluginRecordUsecase) SoftDelete(ctx context.Context, kbID, pluginID, recordType, recordID string) error {
	userID, groupIDs, err := u.viewer(ctx)
	if err != nil {
		return err
	}
	controller, err := u.isKnowledgeBaseController(ctx, kbID, userID)
	if err != nil {
		return err
	}
	record, err := u.records.GetByID(ctx, kbID, pluginID, recordType, recordID, false)
	if err != nil {
		return err
	}
	if !record.CanEdit(userID, controller, groupIDs) {
		return domain.ErrPermissionDenied
	}
	return u.records.SoftDelete(ctx, kbID, pluginID, recordType, recordID)
}

func (u *PluginRecordUsecase) Restore(ctx context.Context, kbID, pluginID, recordType, recordID string) error {
	userID, groupIDs, err := u.viewer(ctx)
	if err != nil {
		return err
	}
	controller, err := u.isKnowledgeBaseController(ctx, kbID, userID)
	if err != nil {
		return err
	}
	record, err := u.records.GetByID(ctx, kbID, pluginID, recordType, recordID, true)
	if err != nil {
		return err
	}
	if !record.DeletedAt.Valid || !record.CanEdit(userID, controller, groupIDs) {
		return domain.ErrPermissionDenied
	}
	return u.records.Restore(ctx, kbID, pluginID, recordType, recordID)
}

func (u *PluginRecordUsecase) validateWrite(ctx context.Context, record *domain.PluginRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	if record.Access.Visibility == domain.PluginRecordVisibilityGroups {
		return u.groups.ValidateIDsBelongToKnowledgeBase(ctx, record.KBID, record.Access.SharedGroupIDs)
	}
	return nil
}

func (u *PluginRecordUsecase) validateGroupMembers(ctx context.Context, group *domain.PluginRecordGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}
	for _, memberUserID := range group.MemberUserIDs {
		allowed, err := u.accessRepo.ValidateKBPerm(group.KBID, memberUserID, consts.UserKBPermissionNotNull)
		if err != nil {
			return fmt.Errorf("validate plugin record group member: %w", err)
		}
		if !allowed {
			return fmt.Errorf("plugin record group member does not have knowledge base access")
		}
	}
	return nil
}

func (u *PluginRecordUsecase) viewer(ctx context.Context) (string, []int, error) {
	auth := domain.GetAuthInfoFromCtx(ctx)
	if auth == nil || strings.TrimSpace(auth.UserId) == "" {
		return "", nil, fmt.Errorf("authenticated user is required")
	}
	groupIDs, err := u.groups.ListIDsForUser(ctx, auth.UserId)
	if err != nil {
		return "", nil, err
	}
	return auth.UserId, groupIDs, nil
}

func (u *PluginRecordUsecase) isKnowledgeBaseController(ctx context.Context, kbID, userID string) (bool, error) {
	auth := domain.GetAuthInfoFromCtx(ctx)
	if auth != nil && auth.IsToken && auth.Permission == consts.UserKBPermissionFullControl {
		return true, nil
	}
	return u.accessRepo.ValidateKBPerm(kbID, userID, consts.UserKBPermissionFullControl)
}
