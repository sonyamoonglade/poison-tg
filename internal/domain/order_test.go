package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string {
	return &s
}

func TestNewOrder(t *testing.T) {
	customer1 := Customer{
		TelegramID:  123,
		Username:    stringPtr("john"),
		FullName:    stringPtr("John Doe"),
		PhoneNumber: stringPtr("123456789"),
		TgState:     State{V: 1},
		Cart: Cart{
			Position{
				ShopLink:  "example.com",
				PriceRUB:  100,
				PriceYUAN: 10,
				Button:    Button95,
				Size:      "xl",
			},
			Position{
				ShopLink:  "example.com",
				PriceRUB:  200,
				PriceYUAN: 20,
				Button:    ButtonGrey,
				Size:      "L",
			},
		},
	}

	tests := []struct {
		description     string
		customer        Customer
		deliveryAddress string
		expectedOrder   Order
	}{
		{
			description:     "test with empty deliveryAddress",
			customer:        customer1,
			deliveryAddress: "",
			expectedOrder: Order{
				Customer:        customer1,
				Cart:            customer1.Cart,
				AmountRUB:       300,
				AmountYUAN:      30,
				DeliveryAddress: "",
				IsPaid:          false,
				IsApproved:      false,
				Status:          StatusNotApproved,
			},
		},
		{
			description:     "test with non-empty deliveryAddress",
			customer:        customer1,
			deliveryAddress: "123 Main St., Anytown, USA",
			expectedOrder: Order{
				Customer:        customer1,
				Cart:            customer1.Cart,
				AmountRUB:       300,
				AmountYUAN:      30,
				DeliveryAddress: "123 Main St., Anytown, USA",
				IsPaid:          false,
				IsApproved:      false,
				Status:          StatusNotApproved,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			order := NewOrder(test.customer, test.deliveryAddress, false, "abcd")
			require.Equal(t, test.expectedOrder.Status, order.Status)
			require.Equal(t, test.expectedOrder.AmountRUB, order.AmountRUB)
			require.Equal(t, test.expectedOrder.AmountYUAN, order.AmountYUAN)
			require.Equal(t, test.expectedOrder.DeliveryAddress, order.DeliveryAddress)
			require.Equal(t, test.expectedOrder.Cart, order.Cart)
			require.False(t, order.IsPaid)
			require.False(t, order.IsApproved)
		})
	}
}

func TestConvertYuan(t *testing.T) {
	tests := []struct {
		description string
		args        ConvertYuanArgs
		expected    uint64
	}{
		{
			description: "expr izh other",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationIZH,
				Category:  CategoryOther,
			},
			expected: expressfn(othMul, 764)(100, 1.0),
		},
		{
			description: "expr spb other",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationSPB,
				Category:  CategoryOther,
			},
			expected: expressfn(othMul, 764)(100, 1.0),
		},
		{
			description: "expr other other",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationOther,
				Category:  CategoryOther,
			},
			expected: expressfn(othMul, 764)(100, 1.0),
		},

		{
			description: "expr izh light",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationIZH,
				Category:  CategoryLight,
			},
			expected: expressfn(lightMul, 764)(100, 1.0),
		},

		{
			description: "expr spb light",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationSPB,
				Category:  CategoryLight,
			},
			expected: expressfn(lightMul, 764)(100, 1.0),
		},
		{
			description: "expr other light",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationOther,
				Category:  CategoryLight,
			},
			expected: expressfn(lightMul, 764)(100, 1.0),
		},
		{
			description: "expr izh heavy",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationOther,
				Category:  CategoryHeavy,
			},
			expected: expressfn(heavyMul, 764)(100, 1.0),
		},
		{
			description: "expr spb heavy",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationSPB,
				Category:  CategoryHeavy,
			},
			expected: expressfn(heavyMul, 764)(100, 1.0),
		},
		{
			description: "expr other heavy",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeExpress,
				Location:  LocationOther,
				Category:  CategoryHeavy,
			},
			expected: expressfn(heavyMul, 764)(100, 1.0),
		},
		{
			description: "normal izh other",
			args: ConvertYuanArgs{
				X:         100,
				Rate:      1.0,
				OrderType: OrderTypeNormal,
				Location:  LocationIZH,
				Category:  CategoryOther,
			},
			expected: normalfn(othMul, 1075)(100, 1.0),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := ConvertYuan(test.args)
			require.Equal(t, test.expected, actual)
		})
	}
}
