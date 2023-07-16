package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dtm-labs/dtmgrpc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/status"
	"mall/service/order/model"
	"mall/service/order/rpc/internal/svc"
	"mall/service/order/rpc/order"
	"mall/service/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateLogic) Create(in *order.CreateRequest) (*order.CreateResponse, error) {
	/*	不用dtm的逻辑
		// 查询用户是否存在
		_, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &user.UserInfoRequest{
			Id: in.Uid,
		})
		if err != nil {
			return nil, err
		}

		// 查询产品是否存在
		productRes, err := l.svcCtx.ProductRpc.Detail(l.ctx, &product.DetailRequest{
			Id: in.Pid,
		})
		if err != nil {
			return nil, err
		}
		// 判断产品库存是否充足
		if productRes.Stock <= 0 {
			return nil, status.Error(500, "产品库存不足")
		}

		newOrder := model.Order{
			Uid:    in.Uid,
			Pid:    in.Pid,
			Amount: in.Amount,
			Status: 0,
		}
		// 创建订单
		res, err := l.svcCtx.OrderModel.Insert(l.ctx, &newOrder)
		if err != nil {
			return nil, status.Error(500, err.Error())
		}

		newOrder.Id, err = res.LastInsertId()
		if err != nil {
			return nil, status.Error(500, err.Error())
		}
		// 更新产品库存
		_, err = l.svcCtx.ProductRpc.Update(l.ctx, &product.UpdateRequest{
			Id:     productRes.Id,
			Name:   productRes.Name,
			Desc:   productRes.Desc,
			Stock:  productRes.Stock - 1,
			Amount: productRes.Amount,
			Status: productRes.Status,
		})
		if err != nil {
			return nil, err
		}

		return &order.CreateResponse{
			Id: newOrder.Id,
		}, nil
		获取 RawDB*/
	db, err := sqlx.NewMysql(l.svcCtx.Config.Mysql.DataSource).RawDB()
	if err != nil {
		return nil, status.Error(500, err.Error())
	}

	// 获取子事务屏障对象
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
	if err != nil {
		return nil, status.Error(500, err.Error())
	}
	// 开启子事务屏障
	if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {
		// 查询用户是否存在
		_, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &user.UserInfoRequest{
			Id: in.Uid,
		})
		if err != nil {
			return fmt.Errorf("用户不存在")
		}

		newOrder := model.Order{
			Uid:    in.Uid,
			Pid:    in.Pid,
			Amount: in.Amount,
			Status: 0,
		}
		// 创建订单
		_, err = l.svcCtx.OrderModel.TxInsert(tx, &newOrder)
		if err != nil {
			return fmt.Errorf("订单创建失败")
		}

		return nil
	}); err != nil {
		return nil, status.Error(500, err.Error())
	}

	return &order.CreateResponse{}, nil
}
