#!/bin/bash

# Script to create a new domain in the gin project
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

echo "Creating domain: ${DOMAIN_NAME_UPPER}"
echo "Directory: ${DOMAIN_DIR}"

# Create domain directory
mkdir -p "$DOMAIN_DIR"

# Create model.go
cat > "$DOMAIN_DIR/model.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import "time"

type ${DOMAIN_NAME_UPPER} struct {
	Id        int64     \`json:"id" gorm:"primaryKey;column:id"\`
	Name      string    \`json:"name" gorm:"column:name"\`
	CreatedAt time.Time \`json:"created_at" gorm:"column:created_at;autoCreateTime"\`
}
EOF

# Create repository.go
cat > "$DOMAIN_DIR/repository.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"context"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/base"
)

type ${DOMAIN_NAME_UPPER}Repository interface {
	GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error)
	GetAll(ctx context.Context) []${DOMAIN_NAME_UPPER}
	Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error)
	RefreshDynamicColumnsById(ctx context.Context, id int64, columnNames []string) error
}

type ${DOMAIN_NAME_LOWER}Repository struct {
	base.BaseHelper
	dynamiccolumnRepo dynamiccolumn.DynamicColumnRepository
}

func New${DOMAIN_NAME_UPPER}Repository(dynamiccolumnRepo dynamiccolumn.DynamicColumnRepository) ${DOMAIN_NAME_UPPER}Repository {
	return &${DOMAIN_NAME_LOWER}Repository{dynamiccolumnRepo: dynamiccolumnRepo}
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

func (r *${DOMAIN_NAME_LOWER}Repository) RefreshDynamicColumnsById(ctx context.Context, id int64, columnNames []string) error {
	err := r.dynamiccolumnRepo.RefreshDynamicColumnsById(ctx, "${DOMAIN_NAME_PLURAL}", id, columnNames)
	return err
}
EOF

# Create service.go
cat > "$DOMAIN_DIR/service.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"context"
)

type ${DOMAIN_NAME_UPPER}Service interface {
	GetAll${DOMAIN_NAME_UPPER}s(ctx context.Context) []${DOMAIN_NAME_UPPER}
	GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error)
	Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error)
}

type ${DOMAIN_NAME_LOWER}Service struct {
	${DOMAIN_NAME_LOWER}Repo ${DOMAIN_NAME_UPPER}Repository
}

func New${DOMAIN_NAME_UPPER}Service(${DOMAIN_NAME_LOWER}Repo ${DOMAIN_NAME_UPPER}Repository) ${DOMAIN_NAME_UPPER}Service {
	return &${DOMAIN_NAME_LOWER}Service{${DOMAIN_NAME_LOWER}Repo: ${DOMAIN_NAME_LOWER}Repo}
}

func (s *${DOMAIN_NAME_LOWER}Service) GetAll${DOMAIN_NAME_UPPER}s(ctx context.Context) []${DOMAIN_NAME_UPPER} {
	return s.${DOMAIN_NAME_LOWER}Repo.GetAll(ctx)
}

func (s *${DOMAIN_NAME_LOWER}Service) GetById(ctx context.Context, id int64) (*${DOMAIN_NAME_UPPER}, error) {
	return s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, id)
}

func (s *${DOMAIN_NAME_LOWER}Service) Create(ctx context.Context, entity *${DOMAIN_NAME_UPPER}) (*${DOMAIN_NAME_UPPER}, error) {
	created, err := s.${DOMAIN_NAME_LOWER}Repo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	
	// Refresh dynamic columns
	err = s.${DOMAIN_NAME_LOWER}Repo.RefreshDynamicColumnsById(ctx, created.Id, nil)
	if err != nil {
		return nil, err
	}
	
	// Fetch updated record
	return s.${DOMAIN_NAME_LOWER}Repo.GetById(ctx, created.Id)
}
EOF

# Create handler.go
cat > "$DOMAIN_DIR/handler.go" <<EOF
package ${DOMAIN_NAME_LOWER}

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/models"
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
	entities := h.${DOMAIN_NAME_LOWER}Service.GetAll${DOMAIN_NAME_UPPER}s(c.Request.Context())
	c.JSON(200, models.NewListResponse(entities, nil, ""))
}

func (h *${DOMAIN_NAME_LOWER}Handler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID"))
		return
	}
	
	entity, err := h.${DOMAIN_NAME_LOWER}Service.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, models.NewErrorResponse("Not found"))
		return
	}
	
	c.JSON(200, models.NewSuccessResponse(entity, ""))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Create(c *gin.Context) {
	var entity ${DOMAIN_NAME_UPPER}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, models.NewErrorResponse(err.Error()))
		return
	}
	
	created, err := h.${DOMAIN_NAME_LOWER}Service.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, models.NewErrorResponse(err.Error()))
		return
	}
	
	c.JSON(201, models.NewSuccessResponse(created, "Created successfully"))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Update(c *gin.Context) {
	c.JSON(501, models.NewErrorResponse("Not implemented"))
}

func (h *${DOMAIN_NAME_LOWER}Handler) Delete(c *gin.Context) {
	c.JSON(501, models.NewErrorResponse("Not implemented"))
}
EOF

# Create route.go
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

# Add import
sed -i.tmp "/\"gin-demo\/internal\/domain\/invoice\"/a\\
	\"gin-demo/internal/domain/${DOMAIN_NAME_LOWER}\"
" "$CONTAINER_FILE"

# Add to Container struct (after Company domain)
sed -i.tmp "/CompanyHandler    company.CompanyHandler/a\\
\\
	// ${DOMAIN_NAME_UPPER} Domain\\
	${DOMAIN_NAME_UPPER}Repository ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Repository\\
	${DOMAIN_NAME_UPPER}Service    ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Service\\
	${DOMAIN_NAME_UPPER}Handler    ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}Handler
" "$CONTAINER_FILE"

# Add to NewModelsMap
sed -i.tmp "/\"companies\": company.Company{},/a\\
		\"${DOMAIN_NAME_PLURAL}\": ${DOMAIN_NAME_LOWER}.${DOMAIN_NAME_UPPER}{},
" "$CONTAINER_FILE"

# Add to NewContainer (after Company initialization)
sed -i.tmp "/c.CompanyHandler = company.NewCompanyHandler(c.CompanyService)/a\\
\\
	// ${DOMAIN_NAME_UPPER}\\
	c.${DOMAIN_NAME_UPPER}Repository = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Repository(c.DynamicColumnRepository)\\
	c.${DOMAIN_NAME_UPPER}Service = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Service(c.${DOMAIN_NAME_UPPER}Repository)\\
	c.${DOMAIN_NAME_UPPER}Handler = ${DOMAIN_NAME_LOWER}.New${DOMAIN_NAME_UPPER}Handler(c.${DOMAIN_NAME_UPPER}Service)
" "$CONTAINER_FILE"

# Clean up temp files
rm -f "$CONTAINER_FILE.tmp"

echo "✓ Container.go updated successfully!"
echo ""
echo "Next steps:"
echo "1. Run: go mod tidy"
echo "2. Add routes in cmd/server/setup.go:"
echo "   ${DOMAIN_NAME_LOWER}.RegisterRoutes(\"v1\", app, []base.HandlerConfig{"
echo "       {Method: \"GET\", Path: \"\", Handler: c.${DOMAIN_NAME_UPPER}Handler.GetAll},"
echo "       {Method: \"GET\", Path: \"/:id\", Handler: c.${DOMAIN_NAME_UPPER}Handler.GetById},"
echo "       {Method: \"POST\", Path: \"\", Handler: c.${DOMAIN_NAME_UPPER}Handler.Create},"
echo "   })"
echo "3. Create database migration for '${DOMAIN_NAME_PLURAL}' table"
echo ""
echo "Domain '${DOMAIN_NAME_UPPER}' created successfully!"
