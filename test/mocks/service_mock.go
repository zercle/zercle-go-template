package mocks

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/dto"
	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

func (m *MockUserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", ctx, req)
	ret0, _ := ret[0].(*dto.UserResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserServiceMockRecorder) Register(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockUserService)(nil).Register), ctx, req)
}

func (m *MockUserService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, req)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserServiceMockRecorder) Login(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, req)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfile", ctx, userID)
	ret0, _ := ret[0].(*dto.UserResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserServiceMockRecorder) GetProfile(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfile", reflect.TypeOf((*MockUserService)(nil).GetProfile), ctx, userID)
}

// MockPostService is a mock of PostService interface.
type MockPostService struct {
	ctrl     *gomock.Controller
	recorder *MockPostServiceMockRecorder
}

type MockPostServiceMockRecorder struct {
	mock *MockPostService
}

func NewMockPostService(ctrl *gomock.Controller) *MockPostService {
	mock := &MockPostService{ctrl: ctrl}
	mock.recorder = &MockPostServiceMockRecorder{mock}
	return mock
}

func (m *MockPostService) EXPECT() *MockPostServiceMockRecorder {
	return m.recorder
}

func (m *MockPostService) CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePost", ctx, userID, req)
	ret0, _ := ret[0].(*dto.PostResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPostServiceMockRecorder) CreatePost(ctx, userID, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePost", reflect.TypeOf((*MockPostService)(nil).CreatePost), ctx, userID, req)
}

func (m *MockPostService) GetPost(ctx context.Context, postID uuid.UUID) (*dto.PostResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPost", ctx, postID)
	ret0, _ := ret[0].(*dto.PostResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPostServiceMockRecorder) GetPost(ctx, postID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPost", reflect.TypeOf((*MockPostService)(nil).GetPost), ctx, postID)
}

func (m *MockPostService) ListPosts(ctx context.Context) ([]*dto.PostResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPosts", ctx)
	ret0, _ := ret[0].([]*dto.PostResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPostServiceMockRecorder) ListPosts(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPosts", reflect.TypeOf((*MockPostService)(nil).ListPosts), ctx)
}
