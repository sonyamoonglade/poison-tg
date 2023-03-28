package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sonyamoonglade/poison-tg/internal/api/input"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	catalogRepo  repositories.Catalog
	orderRepo    repositories.Order
	customerRepo repositories.Customer
	rateProvider *RateProvider
}

type RateProvider struct {
	mu       *sync.RWMutex
	CurrRate float64
}

func NewRateProvider() *RateProvider {
	return &RateProvider{
		mu:       new(sync.RWMutex),
		CurrRate: 11.96,
	}
}

func (r *RateProvider) GetYuanRate() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.CurrRate
}
func (r *RateProvider) UpdateRate(rate float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrRate = rate
}

func NewHandler(catalogRepo repositories.Catalog, orderRepo repositories.Order, customerRepo repositories.Customer, provider *RateProvider) *Handler {
	return &Handler{
		catalogRepo:  catalogRepo,
		rateProvider: provider,
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	api := router.Group("/api")

	api.Post("/updateRate", h.updateRate)
	api.Get("/currentRate", h.currentRate)
	order := api.Group("/order")
	{
		order.Put("/addComment", h.addCommentToOrder)
		order.Post("/delete/:orderId", h.delete)
		order.Put("/approve/:orderId", h.approve)
		order.Put("/changeStatus", h.changeOrderStatus)
		order.Get("/all", h.getAllOrders)
		order.Get("/:shortId", h.getOrderByID)
	}

	catalog := api.Group("/catalog")
	{
		catalog.Get("/all", h.catalog)
		catalog.Post("/addItem", h.addItemToCatalog)
		catalog.Post("/deleteItem", h.removeItemFromCatalog)
		catalog.Put("/rankUp", h.rankUp)
		catalog.Put("/rankDown", h.rankDown)
	}
}

func (h *Handler) updateRate(c *fiber.Ctx) error {
	newRate := c.QueryFloat("rate", 0.0)
	if newRate == 0.0 {
		return fmt.Errorf("empty rate")
	}
	h.rateProvider.UpdateRate(newRate)
	return c.SendStatus(http.StatusOK)
}

func (h *Handler) addCommentToOrder(c *fiber.Ctx) error {
	var inp input.AddCommentToOrderInput
	if err := c.BodyParser(&inp); err != nil {
		return fmt.Errorf("body parsing error: %w", err)
	}

	newOrder, err := h.orderRepo.AddComment(c.Context(), inp.ToDTO())
	if err != nil {
		return fmt.Errorf("can't add comment: %w", err)
	}

	return c.Status(http.StatusOK).JSON(newOrder)
}

func (h *Handler) changeOrderStatus(c *fiber.Ctx) error {
	var inp input.ChangeOrderStatusInput
	if err := c.BodyParser(&inp); err != nil {
		return err
	}
	if ok := domain.IsValidOrderStatus(domain.Status(inp.NewStatus)); !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid status value",
		})
	}
	newOrder, err := h.orderRepo.ChangeStatus(c.Context(), inp.ToDTO())
	if err != nil {
		return fmt.Errorf("can't change status: %w", err)
	}

	return c.Status(http.StatusOK).JSON(newOrder)
}

func (h *Handler) getAllOrders(c *fiber.Ctx) error {
	orders, err := h.orderRepo.GetAll(c.Context())
	if err != nil {
		return fmt.Errorf("get all orders: %w", err)
	}

	return c.Status(http.StatusOK).JSON(orders)
}

func (h *Handler) addItemToCatalog(c *fiber.Ctx) error {
	var inp input.AddItemToCatalogInput
	if err := c.BodyParser(&inp); err != nil {
		return err
	}
	catalog, err := h.catalogRepo.GetCatalog(c.Context())
	if err != nil {
		return fmt.Errorf("get last rank: %w", err)
	}

	rank := len(catalog)

	if err := h.catalogRepo.AddItem(c.Context(), inp.ToNewCatalogItem(uint(rank))); err != nil {
		return fmt.Errorf("add item to catalog: %w", err)
	}

	if err := h.customerRepo.NullifyCatalogOffsets(c.Context()); err != nil {
		return fmt.Errorf("nullify customer catalog offsets: %w", err)
	}

	return h.withNewCatalog(c)
}
func (h *Handler) removeItemFromCatalog(c *fiber.Ctx) error {
	var inp input.RemoveItemFromCatalogInput
	if err := c.BodyParser(&inp); err != nil {
		return err
	}

	if err := h.catalogRepo.RemoveItem(c.Context(), inp.ItemID); err != nil {
		return fmt.Errorf("remove item: %w", err)
	}

	if err := h.customerRepo.NullifyCatalogOffsets(c.Context()); err != nil {
		return fmt.Errorf("nullify customer catalog offsets: %w", err)
	}

	return h.withNewCatalog(c)
}

func (h *Handler) rankUp(c *fiber.Ctx) error {
	var inp input.RankUpInput
	if err := c.BodyParser(&inp); err != nil {
		return err
	}
	// get itemID of current rank
	currentItemRank, err := h.catalogRepo.GetRankByID(c.Context(), inp.ItemID)
	if err != nil {
		return fmt.Errorf("get rank by id: %w", err)
	}

	wantedRank := currentItemRank + 1
	rankDownItemID, err := h.catalogRepo.GetIDByRank(c.Context(), wantedRank)
	if err != nil {
		return fmt.Errorf("get id by rank: %w", err)
	}

	// We should rank up item under inp.ItemID and
	// rank down item under rankDownItemID - otherwise perform a swap

	if err := h.catalogRepo.UpdateRanks(c.Context(), dto.UpdateItemDTO{
		RankUPItemID:   inp.ItemID,
		RankDownItemID: rankDownItemID,
	}); err != nil {
		return fmt.Errorf("update ranks: %w", err)
	}

	return h.withNewCatalog(c)
}

func (h *Handler) rankDown(c *fiber.Ctx) error {
	var inp input.RankDownInput
	if err := c.BodyParser(&inp); err != nil {
		return err
	}

	// get itemID of current rank
	currentItemRank, err := h.catalogRepo.GetRankByID(c.Context(), inp.ItemID)
	if err != nil {
		return fmt.Errorf("get rank by id: %w", err)
	}

	wantedRank := currentItemRank - 1
	rankUpItemID, err := h.catalogRepo.GetIDByRank(c.Context(), wantedRank)
	if err != nil {
		return fmt.Errorf("get id by rank: %w", err)
	}

	// We should rank down item under inp.ItemID and
	// rank up item under rankUpItemID - otherwise perform a swap

	if err := h.catalogRepo.UpdateRanks(c.Context(), dto.UpdateItemDTO{
		RankUPItemID:   rankUpItemID,
		RankDownItemID: inp.ItemID,
	}); err != nil {
		return fmt.Errorf("update ranks: %w", err)
	}

	return h.withNewCatalog(c)
}

func (h *Handler) withNewCatalog(c *fiber.Ctx) error {
	newCatalog, err := h.catalogRepo.GetCatalog(c.Context())
	if err != nil {
		return fmt.Errorf("get catalog: %w", err)
	}
	return c.Status(http.StatusCreated).JSON(newCatalog)
}

func (h *Handler) catalog(c *fiber.Ctx) error {
	newCatalog, err := h.catalogRepo.GetCatalog(c.Context())
	if err != nil {
		return fmt.Errorf("get catalog: %w", err)
	}
	return c.Status(http.StatusOK).JSON(newCatalog)
}

func (h *Handler) getOrderByID(c *fiber.Ctx) error {
	shortId := c.Params("shortId", "")
	if shortId == "" {
		return fmt.Errorf("invalid shortId")
	}

	order, err := h.orderRepo.GetByShortID(c.Context(), shortId)
	if err != nil {
		return fmt.Errorf("get by short id: %w", err)
	}

	return c.Status(http.StatusOK).JSON(order)
}

func (h *Handler) approve(c *fiber.Ctx) error {
	orderId := c.Params("orderId", "")
	id, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return fmt.Errorf("invalid orderId: %w", err)
	}
	order, err := h.orderRepo.Approve(c.Context(), id)
	if err != nil {
		return fmt.Errorf("approve: %w", err)
	}
	return c.Status(http.StatusOK).JSON(order)
}

func (h *Handler) delete(c *fiber.Ctx) error {
	orderId := c.Params("orderId", "")
	id, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return fmt.Errorf("invalid orderId: %w", err)
	}
	if err := h.orderRepo.Delete(c.Context(), id); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return c.SendStatus(http.StatusOK)
}

func (h *Handler) currentRate(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"rate": h.rateProvider.GetYuanRate(),
	})
}
