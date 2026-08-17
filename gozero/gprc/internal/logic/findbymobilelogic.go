package logic

import (
	"context"

	"grpc-demo/internal/svc"
	"grpc-demo/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindByMobileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindByMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByMobileLogic {
	return &FindByMobileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindByMobileLogic) FindByMobile(in *user.FindByMobileRequest) (*user.FindByMobileResponse, error) {
	// todo: add your logic here and delete this line

	return &user.FindByMobileResponse{}, nil
}
