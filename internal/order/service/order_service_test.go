package service

import (
	"context"
	"errors"
	"testing"
	"time"

	invdto "github.com/quangdangfit/goticket/internal/inventory/dto"
	"github.com/quangdangfit/goticket/internal/order/dto"
	"github.com/quangdangfit/goticket/internal/order/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/idempotency"
)

const (
	stID = "stxxxxxxxxxxxxxxxxxxxxxxxx"
	ttID = "ttxxxxxxxxxxxxxxxxxxxxxxxx"
)

type fakeRepo struct {
	orders map[string]*model.Order
	idem   map[string]string // userID|key -> orderID
}

func newRepo() *fakeRepo {
	return &fakeRepo{orders: map[string]*model.Order{}, idem: map[string]string{}}
}
func (r *fakeRepo) CreateWithItems(_ context.Context, o *model.Order, items []model.OrderItem, idem *model.IdempotencyKey) error {
	o.Items = items
	r.orders[o.ID] = o
	if idem != nil {
		r.idem[idem.UserID+"|"+idem.Key] = idem.OrderID
	}
	return nil
}
func (r *fakeRepo) GetByID(_ context.Context, id string) (*model.Order, error) {
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, apperr.ErrNotFound
}
func (r *fakeRepo) GetByIdempotency(_ context.Context, uid, k string) (*model.Order, error) {
	id, ok := r.idem[uid+"|"+k]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return r.orders[id], nil
}
func (r *fakeRepo) UpdateStatus(_ context.Context, id string, from, to model.OrderStatus) error {
	o, ok := r.orders[id]
	if !ok {
		return apperr.ErrNotFound
	}
	if o.Status != from {
		return apperr.ErrConflict
	}
	o.Status = to
	return nil
}

type fakeInv struct {
	available bool
	released  []string
}

func (f *fakeInv) Warm(context.Context, string, []invdto.QuotaSpec) error { return nil }
func (f *fakeInv) Hold(_ context.Context, in invdto.HoldInput) (*invdto.Hold, error) {
	if !f.available {
		return nil, apperr.ErrSoldOut
	}
	return &invdto.Hold{ID: "hold-1", UserID: in.UserID, Status: "held", Items: in.Items}, nil
}
func (f *fakeInv) Release(_ context.Context, id string) error {
	f.released = append(f.released, id)
	return nil
}
func (f *fakeInv) Confirm(context.Context, string) error                  { return nil }
func (f *fakeInv) Get(context.Context, string) (*invdto.Hold, error)      { return nil, apperr.ErrNotFound }
func (f *fakeInv) Available(context.Context, string, string) (int, error) { return 100, nil }

type fakePrice struct{}

func (fakePrice) UnitPrice(context.Context, string) (int64, string, error) {
	return 150_000, "VND", nil
}

type guard struct{ replay bool }

func (g *guard) Reserve(_ context.Context, _, _ string, _ time.Duration) error {
	if g.replay {
		return idempotency.ErrReplay
	}
	return nil
}
func (g *guard) Release(context.Context, string, string) error { return nil }

func sampleInput() dto.CheckoutInput {
	return dto.CheckoutInput{
		IdempotencyKey: "abcdefgh",
		ShowtimeID:     stID,
		Items:          []invdto.Item{{ShowtimeID: stID, TicketTypeID: ttID, Quantity: 2}},
	}
}

func TestCheckout_HappyPath(t *testing.T) {
	r := newRepo()
	inv := &fakeInv{available: true}
	svc := New(r, inv, fakePrice{}, nil, &guard{})
	out, err := svc.Checkout(context.Background(), "user-1", sampleInput())
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if out.Order.TotalMinor != 300_000 {
		t.Fatalf("total=%d", out.Order.TotalMinor)
	}
	if len(r.orders) != 1 {
		t.Fatal("order not persisted")
	}
}

func TestCheckout_SoldOut(t *testing.T) {
	r := newRepo()
	inv := &fakeInv{available: false}
	svc := New(r, inv, fakePrice{}, nil, &guard{})
	_, err := svc.Checkout(context.Background(), "user-1", sampleInput())
	if !errors.Is(err, apperr.ErrSoldOut) {
		t.Fatalf("want sold out, got %v", err)
	}
}

func TestCheckout_Replay_ReturnsExistingOrder(t *testing.T) {
	r := newRepo()
	inv := &fakeInv{available: true}
	svc := New(r, inv, fakePrice{}, nil, &guard{})
	first, err := svc.Checkout(context.Background(), "user-1", sampleInput())
	if err != nil {
		t.Fatal(err)
	}
	svc2 := New(r, inv, fakePrice{}, nil, &guard{replay: true})
	second, err := svc2.Checkout(context.Background(), "user-1", sampleInput())
	if err != nil {
		t.Fatal(err)
	}
	if second.Order.ID != first.Order.ID {
		t.Fatalf("replay returned new order %s vs %s", second.Order.ID, first.Order.ID)
	}
}

func TestCancel_ReleasesHold(t *testing.T) {
	r := newRepo()
	inv := &fakeInv{available: true}
	svc := New(r, inv, fakePrice{}, nil, &guard{})
	out, _ := svc.Checkout(context.Background(), "user-1", sampleInput())
	if err := svc.Cancel(context.Background(), "user-1", out.Order.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(inv.released) != 1 {
		t.Fatalf("expected release, got %v", inv.released)
	}
}
