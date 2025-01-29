package repositories

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
	"os"
	"reflect"
	"testing"
)

func TestOrderRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "billing")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	defer func() {
		_ = os.Unsetenv("DB_DATABASE")
	}()

	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	repo := NewOrderRepository(db)

	t.Run("GetById", func(t *testing.T) {
		emailVO, _ := valueobjects.ParseEmailFromString("test@example.com")
		orderId := insertTestOrder(t, db, emailVO.String(), 1, "NEW")

		order, err := repo.GetById(orderId)

		assert.Nil(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, orderId, order.Id())
		assert.Equal(t, emailVO.String(), order.Email())
		assert.Equal(t, 1, order.PlanId())
		assert.Equal(t, "NEW", order.Status().String())
	})

	t.Run("GetByPlanIdAndEmail", func(t *testing.T) {
		planId := 6
		emailVO, _ := valueobjects.ParseEmailFromString("GetByPlanIdAndEmail@example.com")

		firstOrderId := insertTestOrder(t, db, emailVO.String(), planId, "NEW")
		secondOrderId := insertTestOrder(t, db, emailVO.String(), planId, "PAID")
		thirdOrderId := insertTestOrder(t, db, emailVO.String(), planId, "PAID")

		expectedOrders := []dataobjects.Order{
			dataobjects.NewOrder(firstOrderId, emailVO, planId, valueobjects.NewOrderStatus("NEW")),
			dataobjects.NewOrder(secondOrderId, emailVO, planId, valueobjects.NewOrderStatus("PAID")),
			dataobjects.NewOrder(thirdOrderId, emailVO, planId, valueobjects.NewOrderStatus("PAID")),
		}

		actualOrders, actualOrdersLoadingErr := repo.GetByPlanIdAndEmail(planId, emailVO)

		assert.Nil(t, actualOrdersLoadingErr)

		matchedOrders := 0
		for _, expected := range expectedOrders {
			for _, actual := range actualOrders {
				if expected.Id() == actual.Id() {
					if reflect.DeepEqual(expected, actual) {
						matchedOrders++
					}
				}
			}
		}
		assert.Equal(t, len(expectedOrders), matchedOrders)
	})

	t.Run("Create", func(t *testing.T) {
		emailVO, _ := valueobjects.ParseEmailFromString("create@example.com")
		orderStatus := valueobjects.NewOrderStatus("NEW")

		order := dataobjects.NewOrder(0, emailVO, 2, orderStatus)
		orderId, err := repo.Create(order)

		assert.Nil(t, err)
		assert.NotZero(t, orderId)

		createdOrder, err := repo.GetById(orderId)
		assert.Nil(t, err)
		assert.NotNil(t, createdOrder)
		assert.Equal(t, emailVO.String(), createdOrder.Email())
		assert.Equal(t, 2, createdOrder.PlanId())
		assert.Equal(t, "NEW", order.Status().String())
	})

	t.Run("Update", func(t *testing.T) {
		emailVO, _ := valueobjects.ParseEmailFromString("update@example.com")
		orderId := insertTestOrder(t, db, emailVO.String(), 3, "NEW")

		updatedEmailVO, _ := valueobjects.ParseEmailFromString("updated@example.com")
		updatedOrderStatus := valueobjects.NewOrderStatus("PAID")

		updatedOrder := dataobjects.NewOrder(orderId, updatedEmailVO, 4, updatedOrderStatus)
		err := repo.Update(updatedOrder)
		assert.Nil(t, err)

		order, err := repo.GetById(orderId)
		assert.Nil(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, updatedEmailVO.String(), order.Email())
		assert.Equal(t, 4, order.PlanId())
		assert.Equal(t, "PAID", order.Status().String())
	})

	t.Run("Delete", func(t *testing.T) {
		emailVO, _ := valueobjects.ParseEmailFromString("delete@example.com")
		orderId := insertTestOrder(t, db, emailVO.String(), 5, "NEW")

		order, err := repo.GetById(orderId)
		assert.Nil(t, err)
		assert.NotNil(t, order)

		err = repo.Delete(order)
		assert.Nil(t, err)

		order, err = repo.GetById(orderId)
		assert.NotNil(t, err)
	})
}

func insertTestOrder(t *testing.T, db *sql.DB, email string, planId int, status string) int {
	var id int
	resultErr := db.QueryRow(insertOrder, email, planId, status).Scan(&id)

	if resultErr != nil {
		t.Fatalf("Failed to insert order: %v", resultErr)
	}

	return id
}
