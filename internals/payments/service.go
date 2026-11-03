package payments

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lungtav/golang_ecommerce_backend/internals/db"
)

type Service interface {
	PayOrder (ctx context.Context, orderId string, userId string) (string, error)
	HandleChargeSuccess(ctx context.Context, ev ChargeSuccess) error
}

type svc struct {
	queries *db.Queries
	paystack Initializer
	pool *pgxpool.Pool
}

func NewService (queries *db.Queries, paystack Initializer, pool *pgxpool.Pool) Service{
	return &svc{queries: queries, paystack: paystack}
}


var(
	ErrNoOrder = errors.New("order doesnt exist")
	ErrNoUser = errors.New("user doesnt exist")
)

func (s *svc) PayOrder (ctx context.Context, orderId string, userId string) (string, error){
	var orderID pgtype.UUID
	var userID pgtype.UUID

	if err:= orderID.Scan(orderId); err!=nil{
		slog.Error("failed to convert orderid to uuid")
		return "", err
	}
	
	if err:= userID.Scan(userId); err!=nil{
		slog.Error("failed to convert userid to uuid")
		return "", err
	}

	//get user
	user, err:=s.queries.GetUserById(ctx, userID)

	if errors.Is(err, pgx.ErrNoRows){
		return "", ErrNoUser
	}

	if err!=nil{
		slog.Error("get user by id failed", "error", err)
		return "", err
	}
	

	//find the order with id
	order, err:= s.queries.GetOrderById(ctx, db.GetOrderByIdParams{
		ID: orderID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows){
		return "", ErrNoOrder
	}

	if err!=nil{
		slog.Error("get order by id failed", "error", err)
		return "", err
	}

	//create payment
	paymentID, err:= s.queries.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID: orderID,
		AmountInKobo: int64(order.Amount),
	})

	if err!=nil{
		slog.Error("create payment failed", "error", err)
		return "", err
	}

	authorization_url, err:= s.paystack.Initialize(ctx, int64(order.Amount), user.Email, paymentID.String())
	
	if err!=nil{
		slog.Error("paystack initialization failed", "error", err)

		if err := s.queries.MarkPaymentFailed(context.WithoutCancel(ctx), paymentID); err != nil {
			slog.Error("mark payment failed failed", "error", err)
		}
		return "", err
	}

	return authorization_url, nil

}

func (s *svc) HandleChargeSuccess(ctx context.Context, ev ChargeSuccess) error {
	var pid pgtype.UUID
	if err := pid.Scan(ev.Reference); err != nil {
		return nil
	}

	payment, err := s.queries.GetPaymentByID(ctx, pid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		slog.Error("get payment failed", "error", err)
		return err
	}

	if ev.Currency != "NGN" || ev.Amount != payment.AmountInKobo {
		slog.Error("webhook amount mismatch", "payment_id", ev.Reference,
			"expected", payment.AmountInKobo, "got", ev.Amount, "currency", ev.Currency)
		return nil 
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	
	n, err := qtx.MarkPaymentSucceeded(ctx, db.MarkPaymentSucceededParams{
		ID:                pid,
		PaymentMethod:     pgtype.Text{String: ev.Channel, Valid: true},
		ProviderPaymentID: pgtype.Text{String: strconv.FormatInt(ev.ProviderID, 10), Valid: true},
	})
	if err != nil {
		slog.Error("mark payment succeeded failed", "error", err)
		return err
	}
	if n == 0 {
		return nil 
	}

	m, err := qtx.MarkOrderPaid(ctx, payment.OrderID)
	if err != nil {
		slog.Error("mark order paid failed", "error", err)
		return err
	}
	if m == 0 {
		slog.Error("payment succeeded for non-pending order, needs refund", "payment_id", ev.Reference)
	}

	return tx.Commit(ctx)
}