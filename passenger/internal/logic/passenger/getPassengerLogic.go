package passengerlogic

import (
	"context"
	"errors"

	"onlinecar-common/errorx"
	"onlinecar-passenger/internal/model"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPassengerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPassengerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPassengerLogic {
	return &GetPassengerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取乘客信息
func (l *GetPassengerLogic) GetPassenger(in *pb.GetPassengerReq) (*pb.PassengerInfo, error) {
	l.Infof("获取乘客信息, passengerId=%d", in.Id)

	passengerUser, err := l.svcCtx.PassengerUserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.ErrUserNotFound
		}
		l.Errorf("查询乘客失败, passengerId=%d, err=%v", in.Id, err)
		return nil, errorx.ErrServer
	}

	info := &pb.PassengerInfo{
		Id: passengerUser.Id,
	}

	if passengerUser.PassengerName != nil {
		info.PassengerName = *passengerUser.PassengerName
	}
	if passengerUser.PassengerPhone != nil {
		info.PassengerPhone = *passengerUser.PassengerPhone
	}
	if passengerUser.PassengerGender != nil {
		info.PassengerGender = int64(*passengerUser.PassengerGender)
	}
	if passengerUser.State != nil {
		info.State = int64(*passengerUser.State)
	}
	if passengerUser.ProfilePhoto != nil {
		info.ProfilePhoto = *passengerUser.ProfilePhoto
	}

	return info, nil
}
