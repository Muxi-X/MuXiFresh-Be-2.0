package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type fakeFormUserInfoModel struct {
	usermodel.UserInfoModel
	findOneFn func(ctx context.Context, id string) (*usermodel.UserInfo, error)
}

func (f *fakeFormUserInfoModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	return f.findOneFn(ctx, id)
}

type fakeFormClient struct {
	entryformclient.EntryFormClient
	checkCalled  bool
	checkReq     *entryformclient.CheckReq
	updateCalled bool
}

func (f *fakeFormClient) CheckForm(ctx context.Context, in *entryformclient.CheckReq, opts ...grpc.CallOption) (*entryformclient.CheckResp, error) {
	f.checkCalled = true
	f.checkReq = in
	return &entryformclient.CheckResp{}, nil
}

func (f *fakeFormClient) UpdateForm(ctx context.Context, in *entryformclient.CreateReq, opts ...grpc.CallOption) (*entryformclient.CreateResp, error) {
	f.updateCalled = true
	return &entryformclient.CreateResp{}, nil
}

func formCtxWithUser(uid string) context.Context {
	return context.WithValue(context.Background(), ctxData.CtxKeyJwtUserID, uid)
}

func TestCheckForm_MyselfWithEntryForm(t *testing.T) {
	userID := primitive.NewObjectID()
	entryFormID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: entryFormID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: globalKey.Myself})
	if err != nil {
		t.Fatalf("myself with entry form should pass, got err: %v", err)
	}
	fc := svcCtx.FormClient.(*fakeFormClient)
	if !fc.checkCalled || fc.checkReq.EntryFormID != entryFormID.Hex() {
		t.Fatalf("CheckForm should be called with own entry form id, got %+v", fc.checkReq)
	}
}

func TestCheckForm_OwnIdPasses(t *testing.T) {
	userID := primitive.NewObjectID()
	entryFormID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: entryFormID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: entryFormID.Hex()})
	if err != nil {
		t.Fatalf("own entry form id should pass, got err: %v", err)
	}
}

func TestCheckForm_OtherIdRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	entryFormID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: entryFormID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: otherID.Hex()})
	if err == nil || err.Error() != "无权查看该报名表" {
		t.Fatalf("other user entry form should be rejected, got: %v", err)
	}
	fc := svcCtx.FormClient.(*fakeFormClient)
	if fc.checkCalled {
		t.Fatal("CheckForm should not be called for unauthorized access")
	}
}

func TestCheckForm_MyselfWithoutEntryForm(t *testing.T) {
	userID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: globalKey.Myself})
	if err == nil || err.Error() != "尚未提交报名表" {
		t.Fatalf("myself without entry form should be rejected, got: %v", err)
	}
}

func TestCheckForm_AdminCanViewOthers(t *testing.T) {
	adminID := primitive.NewObjectID()
	otherFormID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: adminID, UserType: globalKey.Admin}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(adminID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: otherFormID.Hex()})
	if err != nil {
		t.Fatalf("admin should view any entry form, got err: %v", err)
	}
	fc := svcCtx.FormClient.(*fakeFormClient)
	if !fc.checkCalled || fc.checkReq.EntryFormID != otherFormID.Hex() {
		t.Fatalf("CheckForm should be called with requested entry form id, got %+v", fc.checkReq)
	}
}

func TestCheckForm_SuperAdminCanViewOthers(t *testing.T) {
	superID := primitive.NewObjectID()
	otherFormID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: superID, UserType: globalKey.SuperAdmin}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewCheckFormLogic(formCtxWithUser(superID.Hex()), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: otherFormID.Hex()})
	if err != nil {
		t.Fatalf("super admin should view any entry form, got err: %v", err)
	}
}

func TestCheckForm_EmptyUserId(t *testing.T) {
	svcCtx := &svc.ServiceContext{FormClient: &fakeFormClient{}}
	l := NewCheckFormLogic(formCtxWithUser(""), svcCtx)

	_, err := l.CheckForm(&types.CheckReq{EntryFormID: globalKey.Myself})
	if err == nil || err.Error() != "身份缺失" {
		t.Fatalf("empty user id should be rejected, got: %v", err)
	}
}
