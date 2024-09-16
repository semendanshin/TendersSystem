package handlers

import (
	"github.com/labstack/echo/v4"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"
)

type bidResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	AuthorType string `json:"authorType"`
	AuthorID   string `json:"authorID"`
	Version    int    `json:"version"`
	CreatedAt  string `json:"createdAt"`
}

func modelToBidResponse(b *models.Bid) bidResponse {
	return bidResponse{
		ID:         b.ID.String(),
		Name:       b.Name,
		Status:     b.Status.String(),
		AuthorType: b.AuthorType.String(),
		AuthorID:   b.AuthorID.String(),
		Version:    b.Version,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05"),
	}
}

type bidFeedbackResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

func modelToBidFeedbackResponse(b *models.BidFeedback) bidFeedbackResponse {
	return bidFeedbackResponse{
		ID:          b.ID.String(),
		Description: b.Description,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05"),
	}
}

type BidHandler struct {
	bidUseCase abstraction.BidUseCaseInterface
}

func NewBidHandler(bidUseCase abstraction.BidUseCaseInterface) *BidHandler {
	return &BidHandler{
		bidUseCase: bidUseCase,
	}
}

func (b *BidHandler) Register(g *echo.Group) {
	g = g.Group("/bids")
	g.POST("/new", b.CreateBid)
	g.GET("/my", b.GetMyBids)
	g.GET("/:tenderID/list", b.GetBidsByTenderID)
	g.GET("/:id/status", b.GetBidStatus)
	g.PUT("/:id/status", b.ChangeBidStatus)
	g.PATCH("/:id/edit", b.EditBid)
	g.PUT("/:id/submit_decision", b.SubmitDecision)
	g.PUT("/:id/feedback", b.Feedback)
	g.PUT("/:id/rollback/:version", b.RollbackBid)
	g.GET("/:tenderID/reviews", b.GetReviews)
}

func (b *BidHandler) CreateBid(c echo.Context) error {
	type request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TenderID    string `json:"tenderID"`
		AuthorType  string `json:"authorType"`
		AuthorID    string `json:"authorID"`
	}

	var req request
	if err := c.Bind(&req); err != nil {
		return err
	}

	var input dto.CreateBidDTO
	{
		var err error

		input.Name = req.Name
		input.Description = req.Description

		input.TenderID, err = models.ParseID(req.TenderID)
		if err != nil {
			return err
		}

		input.AuthorType, err = models.NewBidAuthorType(req.AuthorType)
		if err != nil {
			return err
		}

		input.AuthorID, err = models.ParseID(req.AuthorID)
		if err != nil {
			return err
		}
	}

	bid, err := b.bidUseCase.Create(c.Request().Context(), &input)
	if err != nil {
		return err
	}

	return c.JSON(201, modelToBidResponse(&bid))
}

func (b *BidHandler) GetMyBids(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		Limit    int    `query:"limit"`
		Offset   int    `query:"offset"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	var options []abstraction.PaginationOptFunc
	{
		if q.Limit != 0 {
			options = append(options, abstraction.WithLimit(q.Limit))
		}

		if q.Offset != 0 {
			options = append(options, abstraction.WithOffset(q.Offset))
		}
	}

	bids, err := b.bidUseCase.GetMy(c.Request().Context(), q.Username, options...)
	if err != nil {
		return err
	}

	var response []bidResponse
	for _, bid := range bids {
		response = append(response, modelToBidResponse(&bid))
	}

	return c.JSON(200, response)
}

func (b *BidHandler) GetBidsByTenderID(c echo.Context) error {
	type query struct {
		TenderID string `param:"tenderID"`
		Username string `query:"username"`
		Limit    int    `query:"limit"`
		Offset   int    `query:"offset"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	var options []abstraction.PaginationOptFunc
	{
		if q.Limit != 0 {
			options = append(options, abstraction.WithLimit(q.Limit))
		}

		if q.Offset != 0 {
			options = append(options, abstraction.WithOffset(q.Offset))
		}
	}

	tenderID, err := models.ParseID(q.TenderID)
	if err != nil {
		return err
	}

	bids, err := b.bidUseCase.GetByTenderID(c.Request().Context(), tenderID, q.Username, options...)
	if err != nil {
		return err
	}

	var response []bidResponse
	for _, bid := range bids {
		response = append(response, modelToBidResponse(&bid))
	}

	return c.JSON(200, response)
}

func (b *BidHandler) GetBidStatus(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		BidID    string `param:"id"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	status, err := b.bidUseCase.GetStatus(c.Request().Context(), bidID, q.Username)
	if err != nil {
		return err
	}

	return c.JSON(200, status)
}

func (b *BidHandler) ChangeBidStatus(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		BidID    string `param:"id"`
		Status   string `query:"status"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	status, err := models.NewBidStatus(q.Status)
	if err != nil {
		return err
	}

	bid, err := b.bidUseCase.SetStatus(c.Request().Context(), bidID, q.Username, status)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToBidResponse(&bid))
}

func (b *BidHandler) EditBid(c echo.Context) error {
	type query struct {
		Username    string  `query:"username"`
		BidID       string  `param:"id"`
		Name        *string `json:"name;omitempty"`
		Description *string `json:"description;omitempty"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	var input dto.UpdateBidDTO
	{
		input.Name = q.Name
		input.Description = q.Description
	}

	bid, err := b.bidUseCase.Update(c.Request().Context(), bidID, q.Username, &input)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToBidResponse(&bid))
}

func (b *BidHandler) SubmitDecision(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		BidID    string `param:"id"`
		Decision string `query:"decision"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	decision, err := models.NewBidDecisionType(q.Decision)
	if err != nil {
		return err
	}

	bid, err := b.bidUseCase.SubmitDecision(c.Request().Context(), bidID, q.Username, decision)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToBidResponse(&bid))
}

func (b *BidHandler) Feedback(c echo.Context) error {
	type query struct {
		Username    string `query:"username"`
		BidID       string `param:"id"`
		BidFeedback string `query:"bidFeedback"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	bid, err := b.bidUseCase.LeaveFeedback(c.Request().Context(), bidID, q.Username, q.BidFeedback)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToBidResponse(&bid))
}

func (b *BidHandler) RollbackBid(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		BidID    string `param:"id"`
		Version  int    `param:"version"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	bid, err := b.bidUseCase.Rollback(c.Request().Context(), bidID, q.Username, q.Version)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToBidResponse(&bid))
}

func (b *BidHandler) GetReviews(c echo.Context) error {
	type query struct {
		Username          string `query:"username"`
		BidID             string `param:"tenderID"`
		RequesterUsername string `query:"requesterUsername"`
		AuthorUsername    string `query:"authorUsername"`
		Limit             int    `query:"limit"`
		Offset            int    `query:"offset"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	var options []abstraction.PaginationOptFunc
	{
		if q.Limit != 0 {
			options = append(options, abstraction.WithLimit(q.Limit))
		}

		if q.Offset != 0 {
			options = append(options, abstraction.WithOffset(q.Offset))
		}
	}

	bidID, err := models.ParseID(q.BidID)
	if err != nil {
		return err
	}

	bidFeedbacks, err := b.bidUseCase.GetAuthorsFeedback(c.Request().Context(), bidID, q.RequesterUsername, q.AuthorUsername, options...)
	if err != nil {
		return err
	}

	var response []bidFeedbackResponse
	for _, feedback := range bidFeedbacks {
		response = append(response, modelToBidFeedbackResponse(&feedback))
	}

	return c.JSON(200, response)
}
