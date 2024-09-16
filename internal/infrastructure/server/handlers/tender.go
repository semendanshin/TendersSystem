package handlers

import (
	"fmt"
	"strings"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"

	"github.com/labstack/echo/v4"
)

type tenderResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ServiceType string `json:"serviceType"`
	Version     int    `json:"version"`
	CreatedAt   string `json:"createdAt"`
}

func modelToResponse(t *models.Tender) tenderResponse {
	return tenderResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		Status:      t.Status.String(),
		ServiceType: t.ServiceType.String(),
		Version:     t.Version,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05"),
	}
}

type TenderHandler struct {
	tenderUseCase abstraction.TenderUseCaseInterface
}

func NewTenderHandler(tenderUseCase abstraction.TenderUseCaseInterface) *TenderHandler {
	return &TenderHandler{
		tenderUseCase: tenderUseCase,
	}
}

func (t *TenderHandler) Register(g *echo.Group) {
	g = g.Group("/tenders")
	g.GET("", t.GetTenders)
	g.POST("/new", t.CreateTender)
	g.GET("/my", t.GetMyTenders)
	g.GET("/:id/status", t.GetTenderStatus)
	g.PUT("/:id/status", t.ChangeTenderStatus)
	g.PATCH("/:id/edit", t.EditTender)
	g.PUT("/:id/rollback/:version", t.RollbackTender)
}

func (t *TenderHandler) GetTenders(c echo.Context) error {
	// limit, offset - int (optional)
	// serviceType - array of strings (optional)

	type query struct {
		Limit       int      `query:"limit"`
		Offset      int      `query:"offset"`
		ServiceType []string `query:"service_type"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	var options []abstraction.GetTendersOptFunc
	{
		if q.Limit != 0 || q.Offset != 0 {
			options = append(options, abstraction.WithPaginationOptions(&abstraction.PaginationOptions{
				Limit:  q.Limit,
				Offset: q.Offset,
			}))
		}

		if q.ServiceType != nil {
			for _, strType := range q.ServiceType {
				tenderType, err := models.NewTenderType(strings.ToLower(strType))
				if err != nil {
					return err
				}
				options = append(options, abstraction.WithServiceType(tenderType))
			}
		}
	}

	tenders, err := t.tenderUseCase.GetAll(c.Request().Context(), options...)
	if err != nil {
		return err
	}

	var response []tenderResponse
	for _, tender := range tenders {
		response = append(response, modelToResponse(&tender))
	}

	return c.JSON(200, response)
}

func (t *TenderHandler) CreateTender(c echo.Context) error {

	type body struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		ServiceType     string `json:"serviceType"`
		OrganizationID  string `json:"organizationId"`
		CreatorUsername string `json:"creatorUsername"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return err
	}

	var input dto.CreateTenderDTO
	{
		input.Name = b.Name
		input.Description = b.Description

		var err error
		input.ServiceType, err = models.NewTenderType(strings.ToLower(b.ServiceType))
		if err != nil {
			return err
		}

		input.OrganizationID, err = models.ParseID(b.OrganizationID)
		if err != nil {
			return err
		}

		input.CreatorUsername = b.CreatorUsername
	}

	tender, err := t.tenderUseCase.Create(c.Request().Context(), &input)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToResponse(&tender))
}

func (t *TenderHandler) GetMyTenders(c echo.Context) error {
	type query struct {
		Limit    int    `query:"limit"`
		Offset   int    `query:"offset"`
		Username string `query:"username"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return fmt.Errorf("failed to bind query: %w", err)
	}

	var options []abstraction.PaginationOptFunc
	{
		if q.Limit != 0 || q.Offset != 0 {
			options = append(options, abstraction.WithOffset(q.Offset), abstraction.WithLimit(q.Limit))
		}
	}

	tenders, err := t.tenderUseCase.GetMy(c.Request().Context(), q.Username, options...)
	if err != nil {
		return fmt.Errorf("failed to get my tenders: %w", err)
	}

	var response []tenderResponse
	for _, tender := range tenders {
		response = append(response, modelToResponse(&tender))
	}

	return c.JSON(200, response)
}

func (t *TenderHandler) GetTenderStatus(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		TenderID string `param:"id"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	tenderID, err := models.ParseID(q.TenderID)
	if err != nil {
		return err
	}

	status, err := t.tenderUseCase.GetStatus(c.Request().Context(), tenderID, q.Username)
	if err != nil {
		return err
	}

	return c.JSON(200, status.String())
}

func (t *TenderHandler) ChangeTenderStatus(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		TenderID string `param:"id"`
		Status   string `query:"status"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	tenderID, err := models.ParseID(q.TenderID)
	if err != nil {
		return err
	}

	status, err := models.NewTenderStatus(strings.ToLower(q.Status))
	if err != nil {
		return err
	}

	tender, err := t.tenderUseCase.SetStatus(c.Request().Context(), tenderID, q.Username, status)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToResponse(&tender))
}

func (t *TenderHandler) EditTender(c echo.Context) error {
	type query struct {
		Username    string  `query:"username"`
		TenderID    string  `param:"id"`
		Name        *string `json:"name;omitempty"`
		Description *string `json:"description;omitempty"`
		ServiceType *string `json:"serviceType;omitempty"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return c.JSON(400, err)
	}

	tenderID, err := models.ParseID(q.TenderID)
	if err != nil {
		return err
	}

	var input dto.UpdateTenderDTO
	{
		input.Name = q.Name
		input.Description = q.Description
		if q.ServiceType != nil {
			serviceType, err := models.NewTenderType(strings.ToLower(*q.ServiceType))
			if err != nil {
				return err
			}
			input.ServiceType = &serviceType
		}
	}

	tender, err := t.tenderUseCase.Update(c.Request().Context(), tenderID, q.Username, &input)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToResponse(&tender))
}

func (t *TenderHandler) RollbackTender(c echo.Context) error {
	type query struct {
		Username string `query:"username"`
		TenderID string `param:"id"`
		Version  int    `param:"version"`
	}

	var q query
	if err := c.Bind(&q); err != nil {
		return err
	}

	tenderID, err := models.ParseID(q.TenderID)
	if err != nil {
		return err
	}

	tender, err := t.tenderUseCase.Rollback(c.Request().Context(), tenderID, q.Username, q.Version)
	if err != nil {
		return err
	}

	return c.JSON(200, modelToResponse(&tender))
}
