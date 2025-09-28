package handlers

import (
	"net/http"
	"strconv"

	"miningpet/internal/services"
	"github.com/gin-gonic/gin"
)

type PetHandler struct {
	petService *services.PetService
}

func NewPetHandler(petService *services.PetService) *PetHandler {
	return &PetHandler{
		petService: petService,
	}
}

type CreatePetRequest struct {
	OwnerName string `json:"owner_name" binding:"required"`
}

func (h *PetHandler) CreatePet(c *gin.Context) {
	var req CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pet := h.petService.CreatePet(req.OwnerName)
	c.JSON(http.StatusCreated, pet)
}

func (h *PetHandler) GetPet(c *gin.Context) {
	petID := c.Param("id")
	
	pet, exists := h.petService.GetPet(petID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pet not found"})
		return
	}

	c.JSON(http.StatusOK, pet)
}

func (h *PetHandler) GetAllPets(c *gin.Context) {
	pets := h.petService.GetAllPets()
	c.JSON(http.StatusOK, gin.H{"pets": pets})
}

func (h *PetHandler) StartExploration(c *gin.Context) {
	petID := c.Param("id")
	
	err := h.petService.StartExploration(petID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exploration started"})
}

func (h *PetHandler) GetEvents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	events := h.petService.GetRecentEvents(limit)
	c.JSON(http.StatusOK, gin.H{"events": events})
}