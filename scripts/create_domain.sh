#!/bin/bash

# Script to create a new domain in the gin project based on invoice template
# Usage: ./scripts/create_domain.sh <domain_name>

set -e

if [ -z "$1" ]; then
    echo "Usage: ./scripts/create_domain.sh <domain_name>"
    echo "Example: ./scripts/create_domain.sh product"
    exit 1
fi

DOMAIN_NAME=$1
DOMAIN_NAME_LOWER=$(echo "$DOMAIN_NAME" | tr '[:upper:]' '[:lower:]')
DOMAIN_NAME_UPPER=$(echo "$DOMAIN_NAME" | awk '{print toupper(substr($0,1,1)) tolower(substr($0,2))}')
DOMAIN_NAME_PLURAL="${DOMAIN_NAME_LOWER}s"

DOMAIN_DIR="internal/domain/${DOMAIN_NAME_LOWER}"
SOURCE_DIR="internal/domain/invoice"

echo "Creating domain: ${DOMAIN_NAME_UPPER} (based on invoice template)"
echo "Directory: ${DOMAIN_DIR}"

# Create domain directory
mkdir -p "$DOMAIN_DIR"

# Copy and transform model.go
echo "Creating model.go..."
cat > "$DOMAIN_DIR/model.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import "gin-demo/internal/shared/types"

type ${DOMAIN_NAME_UPPER} struct {
	types.GormModel
	Name string \`json:"name" gorm:"column:name" binding:"required"\`
}

type ${DOMAIN_NAME_UPPER}UpdateRequest struct {
	Name *string \`json:"name,omitempty"\`
}
EOF

# Copy and transform repository.go
echo "Creating repository.go..."
cat > "$DOMAIN_DIR/repository.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"context"
	"gin-demo/internal/shared/base"
)

type ${DOMAIN_NAME_UPPER}Repository interface {
	GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error)
	GetAll(ctx context.Context) []${DOMAIN_NAME_UPPER}
	Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error)
	Update(ctx context.Context, id int64, updatePayload *${DOMAIN_NAME_UPPER}UpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type ${DOMAIN_NAME_LOWER}Repository struct {
	base.BaseHelper
}

func New${DOMAIN_NAME_UPPER}Repository() ${DOMAIN_NAME_UPPER}Repository {
	return &${DOMAIN_NAME_LOWER}Repository{}
}

func (r *${DOMAIN_NAME_LOWER}Repository) GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error) {
	tx := r.GetDbTx(ctx)
	var entity ${DOMAIN_NAME_UPPER}
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *${DOMAIN_NAME_LOWER}Repository) GetAll(ctx context.Context) []${DOMAIN_NAME_UPPER} {
	tx := r.GetDbTx(ctx)
	var entities []${DOMAIN_NAME_UPPER}
	tx.Find(&entities)
	return entities
}

func (r *${DOMAIN_NAME_LOWER}Repository) Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *${DOMAIN_NAME_LOWER}Repository) Update(ctx context.Context, id int64, updatePayload *${DOMAIN_NAME_UPPER}UpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&${DOMAIN_NAME_UPPER}{}).Where("id = ?", id).Updates(updatePayload).Error
}

func (r *${DOMAIN_NAME_LOWER}Repository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&${DOMAIN_NAME_UPPER}{}, id).Error
}
EOF

# Copy and transform service.go
echo "Creating service.go..."
cat > "$DOMAIN_DIR/service.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"context"
	"errors"
	"gin-demo/internal/system/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type ${DOMAIN_NAME_UPPER}Service interface {
	GetAll(ctx context.Context) []${DOMAIN_NAME_UPPER}
	GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error)
	Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error)
	Update(ctx context.Context, id int64, updatePayload *${DOMAIN_NAME_UPPER}UpdateRequest) (*${DOMAIN_NAME_UPPER}, error)
	Delete(ctx context.Context, id int64) error
}

type ${DOMAIN_NAME_LOWER}Service struct {
	${DOMAIN_NAME_LOWER}Repo      ${DOMAIN_NAME_UPPER}Repository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func New${DOMAIN_NAME_UPPER}Service(${DOMAIN_NAME_LOWER}Repo ${DOMAIN_NAME_UPPER}Repository, dynamicColumnService dynamiccolumn.DynamicColumnService) ${DOMAIN_NAME_UPPER}Service {
	return &${DOMAIN_NAME_LOWER}Service{
		${DOMAIN_NAME_LOWER}Repo:      ${DOMAIN_NAME_LOWER}Repo,
		dynamicColumnService: dynamicColumnService,
	}
}

func (s *${DOMAIN_NAME_LOWER}Service) GetAll(ctx context.Context) []${DOMAIN_NAME_UPPER} {
	return s.${DOMAIN_NAME_LOWER}Repo.GetAll(ctx)
}

func (s *${DOMAIN_NAME_LOWER}Service) GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error) {
	return s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, id)
}

func (s *${DOMAIN_NAME_LOWER}Service) Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error) {
	entity, err := s.${DOMAIN_NAME_LOWER}Repo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	
	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "${DOMAIN_NAME_PLURAL}", []int64{entity.Id}, constants.ActionCreate, nil, nil, entity)
	if err != nil {
		return nil, err
	}
	
	// Fetch updated record with dynamic columns
	refreshedEntity, err := s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, entity.Id)
	if err != nil {
		return nil, err
	}
	
	return refreshedEntity, nil
}

func (s *${DOMAIN_NAME_LOWER}Service) Update(ctx context.Context, id int64, updatePayload *${DOMAIN_NAME_UPPER}UpdateRequest) (*${DOMAIN_NAME_UPPER}, error) {
	originalEntity, err := s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	
	err = s.${DOMAIN_NAME_LOWER}Repo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}
	
	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "${DOMAIN_NAME_PLURAL}", []int64{id}, constants.ActionUpdate, nil, &originalEntity.Id, updatePayload)
	if err != nil {
		return nil, err
	}
	
	// Fetch updated record
	refreshedEntity, err := s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return refreshedEntity, nil
}

func (s *${DOMAIN_NAME_LOWER}Service) Delete(ctx context.Context, id int64) error {
	originalEntity, err := s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalEntity == nil {
		return errors.New("${DOMAIN_NAME_LOWER} not found")
	}
	
	err = s.${DOMAIN_NAME_LOWER}Repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	
	// Refresh dynamic columns after deletion
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "${DOMAIN_NAME_PLURAL}", []int64{id}, constants.ActionDelete, nil, &originalEntity.Id, nil)
	return err
}
EOF

# Copy and transform handler.go
echo "Creating handler.go..."
cat > "$DOMAIN_DIR/handler.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ${DOMAIN_NAME_UPPER}Handler interface {
	base.BaseHandler
}

type ${DOMAIN_NAME_LOWER}Handler struct {
	${DOMAIN_NAME_LOWER}Service ${DOMAIN_NAME_UPPER}Service
}

func New${DOMAIN_NAME_UPPER}Handler(${DOMAIN_NAME_LOWER}Service ${DOMAIN_NAME_UPPER}Service) ${DOMAIN_NAME_UPPER}Handler {
	return &${DOMAIN_NAME_LOWER}Handler{${DOMAIN_NAME_LOWER}Service: ${DOMAIN_NAME_LOWER}Service}
}

func (h *${DOMAIN_NAME_LOWER}Handler) GetAll(c *gin.Context) {
	entities := h.${DOMAIN_NAME_LOWER}Service.GetAll(c.Request.Context())
	c.JSON(200, types.NewListResponse(entities, nil, ""))
}

func (h *${DOMAIN_NAME_LOWER}Handler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	entity, err := h.${DOMAIN_NAME_LOWER}Service.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, types.NewErrorResponse("Not found", err.Error()))
		return
	}
	
	c.JSON(200, types.NewSingleResponse[${DOMAIN_NAME_UPPER}](entity, ""))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Create(c *gin.Context) {
	var entity ${DOMAIN_NAME_UPPER}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}
	
	created, err := h.${DOMAIN_NAME_LOWER}Service.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(201, types.NewSingleResponse[${DOMAIN_NAME_UPPER}](created, "Created successfully"))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	var updatePayload ${DOMAIN_NAME_UPPER}UpdateRequest
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}
	
	updated, err := h.${DOMAIN_NAME_LOWER}Service.Update(c.Request.Context(), id, &updatePayload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(200, types.NewSingleResponse[${DOMAIN_NAME_UPPER}](updated, "Updated successfully"))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	err = h.${DOMAIN_NAME_LOWER}Service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(200, types.NewSingleResponse[${DOMAIN_NAME_UPPER}](nil, "Deleted successfully"))
}
EOF

# Create route.go
echo "Creating route.go..."
cat > "$DOMAIN_DIR/route.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/shared/base"
)

func RegisterRoutes(version string, app *config.App, handlers []base.HandlerConfig) {
	group := app.Group(fmt.Sprintf("/api/%s/${DOMAIN_NAME_PLURAL}", version))
	for _, h := range handlers {
		group.Handle(h.Method, h.Path, h.Handler)
	}
}
EOF

echo ""
echo "✓ Domain files created successfully!"
echo ""
echo "Now updating container.go..."

# Update container.go
CONTAINER_FILE="internal/application/container/container.go"

# Backup container.go
cp "$CONTAINER_FILE" "$CONTAINER_FILE.bak"

# Add import (after invoice domain import)
sed -i.tmp "/\"gin-demo\/internal\/domain\/invoice\"/a\\
	\"gin-demo/internal/domain/${DOMAIN_NAME_LOWER}\"
" "$CONTAINER_FILE"

# Add to Container struct (after Invoice domain)
sed -i.tmp "/InvoiceHandler    invoice.InvoiceHandler/a\\
\\
	// ${DOMAIN_NAME_UPPER} Domain\\
	${DOMAIN_NAME_UPPER}Repository ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Repository\\
	${DOMAIN_NAME_UPPER}Service    ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Service\\
	${DOMAIN_NAME_UPPER}Handler    ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Handler
" "$CONTAINER_FILE"

# Add to NewModelsMap (before the closing brace of the map)
sed -i.tmp "/\"payments\":  payment.Payment{},/a\\
		\"${DOMAIN_NAME_PLURAL}\": ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}{},
" "$CONTAINER_FILE"

# Add to NewContainer (after Invoice initialization)
sed -i.tmp "/c.InvoiceHandler = invoice.NewInvoiceHandler(c.InvoiceService)/a\\
\\
	// ${DOMAIN_NAME_UPPER}\\
	c.${DOMAIN_NAME_UPPER}Repository = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Repository()\\
	c.${DOMAIN_NAME_UPPER}Service = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Service(c.${DOMAIN_NAME_UPPER}Repository, c.DynamicColumnService)\\
	c.${DOMAIN_NAME_UPPER}Handler = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Handler(c.${DOMAIN_NAME_UPPER}Service)
" "$CONTAINER_FILE"

# Clean up temp files
rm -f "$CONTAINER_FILE.tmp"

echo "✓ Container.go updated successfully!"

# Update setup.go to add routes
echo "Updating cmd/server/setup.go to add routes..."
SETUP_FILE="cmd/server/setup.go"

# Backup setup.go
cp "$SETUP_FILE" "$SETUP_FILE.bak"

# Add import (after the last domain import)
sed -i.tmp "/\"gin-demo\/internal\/domain\/payment\"/a\\
	\"gin-demo/internal/domain/${DOMAIN_NAME_LOWER}\"
" "$SETUP_FILE"

# Add routes (before the closing brace of SetupRoutes function)
sed -i.tmp "/payment.RegisterRoutes/a\\
	${DOMAIN_NAME_LOWER}.RegisterRoutes(\"v1\", app, []base.HandlerConfig{\\
		{Method: \"GET\", Path: \"\", Handler: c.${DOMAIN_NAME_UPPER}Handler.GetAll},\\
		{Method: \"GET\", Path: \"/:id\", Handler: c.${DOMAIN_NAME_UPPER}Handler.GetById},\\
		{Method: \"POST\", Path: \"\", Handler: c.${DOMAIN_NAME_UPPER}Handler.Create},\\
		{Method: \"PUT\", Path: \"/:id\", Handler: c.${DOMAIN_NAME_UPPER}Handler.Update},\\
		{Method: \"DELETE\", Path: \"/:id\", Handler: c.${DOMAIN_NAME_UPPER}Handler.Delete},\\
	})
" "$SETUP_FILE"

# Clean up temp files
rm -f "$SETUP_FILE.tmp"

echo "✓ Setup.go updated successfully!"
echo ""
echo "=================================================================================="
echo "✓ Domain '${DOMAIN_NAME_UPPER}' created successfully!"
echo "=================================================================================="
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Run: go mod tidy"
echo ""
echo "2. Update model in ${DOMAIN_DIR}/model.go with your actual fields"
echo ""
echo "3. Create database migration:"
echo "   goose -dir migrations create create_${DOMAIN_NAME_PLURAL}_table sql"
echo ""
echo "4. Example migration SQL:"
echo "   CREATE TABLE ${DOMAIN_NAME_PLURAL} ("
echo "       id         BIGSERIAL PRIMARY KEY,"
echo "       name       VARCHAR(255) NOT NULL,"
echo "       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,"
echo "       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
echo "   );"
echo ""
echo "=================================================================================="
echo ""
echo "Features included:"
echo "  ✓ Full CRUD operations (Create, Read, Update, Delete)"
echo "  ✓ Dynamic column integration with refresh on all operations"
echo "  ✓ Update request with omitempty fields for partial updates"
echo "  ✓ Proper error handling and validation"
echo "  ✓ RESTful HTTP handlers with appropriate status codes"
echo "=================================================================================="
