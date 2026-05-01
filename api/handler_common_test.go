package api

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) CreateProfile(ctx context.Context, user repo.UserInput) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *storageMock) CreateProfileWithCountryCheck(ctx context.Context, user repo.UserInput) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *storageMock) GetProfile(ctx context.Context, keycloakID uuid.UUID) (repo.User, error) {
	args := m.Called(ctx, keycloakID)
	return args.Get(0).(repo.User), args.Error(1)
}

func (m *storageMock) UpdateProfile(ctx context.Context, keycloakID uuid.UUID, user repo.UserInput) error {
	args := m.Called(ctx, keycloakID, user)
	return args.Error(0)
}

func (m *storageMock) SetSpouse(ctx context.Context, keycloakID1, keyclaokID2 uuid.UUID, forceUpdate bool) error {
	args := m.Called(ctx, keycloakID1, keyclaokID2, forceUpdate)
	return args.Error(0)
}

func (m *storageMock) DeleteProfile(ctx context.Context, keycloakID uuid.UUID) error {
	args := m.Called(ctx, keycloakID)
	return args.Error(0)
}

func (m *storageMock) HardDeleteProfile(ctx context.Context, keycloakID uuid.UUID) error {
	args := m.Called(ctx, keycloakID)
	return args.Error(0)
}

func (m *storageMock) GetMultipleProfiles(ctx context.Context, intSkip int, intLimit int, country string, email string,
	name string, keycloakID string, tenGroupName string, language string, firstLanguage string, otherLanguageOne string,
	otherLanguageTwo string, otherLanguageThree string, otherLanguageFour string, updatedAt string, createdAt string,
	membership string, membershipType string, convention string, ticket string, galaxy string, gender string,
	userID string, checkAlternativeEmails bool, clause string) ([]repo.User, error) {
	args := m.Called(ctx, intSkip, intLimit, country, email, name, keycloakID, language, firstLanguage, otherLanguageOne,
		otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership, membershipType, convention,
		ticket, galaxy, gender, userID, checkAlternativeEmails, clause)
	return args.Get(0).([]repo.User), args.Error(0)
}

func (m *storageMock) FetchProfileBasedOnPhoneNumber(ctx context.Context, phoneNumber string) (repo.User, error) {
	args := m.Called(ctx, phoneNumber)
	return args.Get(0).(repo.User), args.Error(1)
}

func (m *storageMock) CreateRequest(ctx context.Context, request repo.NewRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *storageMock) ConcludeRequest(ctx context.Context, reqID int, conclusion repo.RequestConclusion) error {
	args := m.Called(ctx, reqID, conclusion)
	return args.Error(0)
}

func (m *storageMock) GetRequestByID(ctx context.Context, id int) (*repo.NewRequest, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repo.NewRequest), args.Error(1)
}

func (m *storageMock) GetMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string, email string, typeFilters []string, orderByCreatedAt string) ([]repo.RequestAndGrant, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMultipleRequestCount(ctx context.Context, kcid string, status string, name string, email string, typeFilters []string, orderByCreatedAt string) (int, error) {
	panic("implement me")
}

func (m *storageMock) GetMembershipByID(ctx context.Context, id int) (repo.Membership, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMembershipByUserID(ctx context.Context, userID string) (repo.UserMembershipRes, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMembershipByKCID(ctx context.Context, kcID string) (repo.UserMembershipRes, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMultipleMembership(ctx context.Context, intSkip int, intLimit int, userID string) ([]repo.Membership, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetExpiredMemberships(ctx context.Context, intSkip int, intLimit int) ([]repo.Membership, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) PatchMembershipByID(ctx context.Context, membership repo.Membership, id int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) SoftDeleteMembershipByID(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) CancelMembership(ctx context.Context, body repo.EmailKeycloakAndUserIDBody) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (repo.MembershipAutomatic, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) EvaluateMembershipByUserID(ctx context.Context, evalbody repo.EmailKeycloakAndUserIDBody) (repo.UserMembershipRes, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetGrantByID(ctx context.Context, id int) (repo.Grant, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMultipleGrant(ctx context.Context, intSkip int, intLimit int, boolCancelled *bool, userID string, grantType string, createdAt string) ([]repo.Grant, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetNotificationByID(ctx context.Context, id int) (repo.Notification, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) CreateNotification(ctx context.Context, noti repo.Notification) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMultipleNotification(ctx context.Context, intSkip int, intLimit int) ([]repo.Notification, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) PatchNotification(ctx context.Context, noti repo.Notification, id int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) SoftDeleteNotification(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetUserNotificationByID(ctx context.Context, id int) (repo.UserNotification, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) CreateUserNotification(ctx context.Context, noti repo.UserNotification) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetMultipleUserNotification(ctx context.Context, intSkip int, intLimit int) ([]repo.UserNotification, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetActiveUserNotificationByUserID(ctx context.Context, userID string) ([]repo.UserNotification, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) PatchUserNotification(ctx context.Context, noti repo.UserNotification, id int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) SoftDeleteUserNotification(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) PerformOperation(ctx context.Context, opr repo.OperationReq) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) RevertOperation(ctx context.Context, newEmail string, oldEmail string) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) IsSubjectID(ctx context.Context, arg1 string, arg2 string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) MergeAccounts(ctx context.Context, accountsMergeData repo.AccountsMergeData) (*repo.EmailKeycloakAndUserIDBody, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) FetchPageNotes(ctx context.Context, pageId int, pageKeycloakId string) ([]repo.PageNote, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) InsertPageNote(ctx context.Context, pageId int, pageKeycloakId string, authorKeycloackId string, note string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) DeletePageNote(ctx context.Context, noteId int, authorKeycloakId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) Close() {
	//TODO implement me
	panic("implement me")
}

func MakeKeycloakAPIMockFactory(prepMock func(*keycloakAPIMock)) keycloak.KeycloakServiceFactory {
	m := new(keycloakAPIMock)
	prepMock(m)
	return func() keycloak.KeycloakService {
		return m
	}
}

type keycloakAPIMock struct {
	mock.Mock
}

func (m *keycloakAPIMock) UpdateUser(ctx context.Context, keycloakID string, firstName *string, lastName *string) error {
	args := m.Called(ctx, keycloakID, firstName, lastName)
	return args.Error(0)
}
