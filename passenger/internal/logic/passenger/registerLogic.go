package passengerlogic

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"onlinecar-common/errorx"
	"onlinecar-common/validate"
	"onlinecar-passenger/internal/model"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// 注册乘客（事务：检查手机号 + 插入乘客）
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	l.Infof("乘客注册, phone=%s, name=%s", in.PassengerPhone, in.PassengerName)

	// 输入校验
	if !validate.IsValidPhone(in.PassengerPhone) {
		return nil, errorx.ErrPhoneFormat
	}
	if utf8.RuneCountInString(in.PassengerName) < 2 || utf8.RuneCountInString(in.PassengerName) > 20 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "姓名长度需在2-20个字符之间")
	}
	if in.PassengerGender < 0 || in.PassengerGender > 2 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "性别参数无效")
	}

	state := int8(0) // 有效
	gender := int8(in.PassengerGender)

	var passengerId int64

	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 事务内检查手机号是否已注册（避免 TOCTOU 竞态）
		existing, err := l.svcCtx.PassengerUserModel.FindByPhoneWithConn(ctx, session, in.PassengerPhone)
		if err != nil {
			return fmt.Errorf("查询乘客失败: %w", err)
		}
		if existing != nil {
			return errorx.ErrPhoneRegistered
		}

		passengerUser := &model.PassengerUser{
			PassengerPhone:  &in.PassengerPhone,
			PassengerName:   &in.PassengerName,
			PassengerGender: &gender,
			State:           &state,
		}

		result, err := l.svcCtx.PassengerUserModel.InsertWithConn(ctx, session, passengerUser)
		if err != nil {
			return fmt.Errorf("插入乘客失败: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("获取乘客ID失败: %w", err)
		}
		passengerId = id
		return nil
	})
	if err != nil {
		var codeErr *errorx.CodeError
		if errors.As(err, &codeErr) {
			return nil, codeErr
		}
		logx.Errorf("乘客注册事务失败, phone=%s, err=%v", in.PassengerPhone, err)
		return nil, errorx.ErrServer
	}

	l.Infof("乘客注册成功, passengerId=%d", passengerId)
	return &pb.RegisterResp{
		Id:             passengerId,
		PassengerName:  in.PassengerName,
		PassengerPhone: in.PassengerPhone,
		State:          int64(state),
	}, nil
}
