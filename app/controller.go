package app

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	errIDNotFound = fmt.Errorf("id not found")
)

type store struct {
	mu sync.RWMutex
	m  map[uuid.UUID]Risk
}

func NewStore() *store {
	return &store{m: make(map[uuid.UUID]Risk)}
}

func (d *store) put(id uuid.UUID, risk Risk) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[id] = risk
}

func (d *store) get(id uuid.UUID) (Risk, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.m[id]
	return v, ok
}

func (d *store) getAll() []Risk {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// todo: what if it's a large map?
	valList := make([]Risk, len(d.m))
	i := 0
	for _, v := range d.m {
		valList[i] = v
		i++
	}
	return valList
}

type riskController struct {
	store *store
}

func NewRiskController(store *store) *riskController {
	return &riskController{store: store}
}

func (rh *riskController) list() gin.HandlerFunc {
	return func(c *gin.Context) {
		valList := rh.store.getAll()
		c.JSON(http.StatusOK, valList)
	}
}

func (rh *riskController) get() gin.HandlerFunc {
	return func(c *gin.Context) {
		inputID := c.Param("id")
		var id uuid.UUID
		var err error
		if id, err = uuid.Parse(inputID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": errIDNotFound.Error()})
			return
		}

		r, ok := rh.store.get(id)

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": errIDNotFound.Error()})
			return
		}
		c.JSON(http.StatusOK, r)

	}
}

func (rh *riskController) post() gin.HandlerFunc {
	return func(c *gin.Context) {

		var inputRisk CreateRisk
		if err := c.ShouldBindJSON(&inputRisk); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				c.JSON(http.StatusBadRequest,
					gin.H{
						"error": fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", ve[0].Field(), ve[0].Tag()),
					})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newRisk := Risk{
			ID:          uuid.New(),
			State:       inputRisk.State,
			Title:       inputRisk.Title,
			Description: inputRisk.Description,
		}

		rh.store.put(newRisk.ID, newRisk)
		c.JSON(http.StatusCreated, newRisk)

	}
}
