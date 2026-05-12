package domain

type OrderItems struct {
	Quantity uint
	ItemId   uint

	Item  Item
	Order Order
}
