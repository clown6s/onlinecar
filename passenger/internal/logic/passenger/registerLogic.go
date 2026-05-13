package passengerlogic

import (
	"context"
	"fmt"

	"onlinecar-passenger/internal/model"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册乘客
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	l.Infof("乘客注册, phone=%s, name=%s", in.PassengerPhone, in.PassengerName)

	// 检查手机号是否已注册
	existing, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil {
		l.Errorf("查询乘客失败, phone=%s, err=%v", in.PassengerPhone, err)
		return nil, fmt.Errorf("查询失败: %v", err)
	}
	if existing != nil {
		l.Errorf("手机号已注册, phone=%s", in.PassengerPhone)
		return nil, fmt.Errorf("该手机号已注册")
	}

	state := int8(0) // 有效
	gender := int8(in.PassengerGender)

	passengerUser := &model.PassengerUser{
		PassengerPhone:  &in.PassengerPhone,
		PassengerName:   &in.PassengerName,
		PassengerGender: &gender,
		State:           &state,
	}

	result, err := l.svcCtx.PassengerUserModel.Insert(l.ctx, passengerUser)
	if err != nil {
		l.Errorf("插入乘客失败, phone=%s, err=%v", in.PassengerPhone, err)
		return nil, fmt.Errorf("注册失败: %v", err)
	}
	passengerId, _ := result.LastInsertId()

	l.Infof("乘客注册成功, passengerId=%d", passengerId)
	return &pb.RegisterResp{
		Id:             passengerId,
		PassengerName:  in.PassengerName,
		PassengerPhone: in.PassengerPhone,
		State:          int64(state),
	}, nil
}
