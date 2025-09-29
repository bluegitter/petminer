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

	pet, err := h.petService.CreatePet(req.OwnerName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

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

type CommandRequest struct {
	Command string                 `json:"command" binding:"required"`
	Params  map[string]interface{} `json:"params"`
}

// ExecuteCommand 执行宠物指令
func (h *PetHandler) ExecuteCommand(c *gin.Context) {
	petID := c.Param("id")
	
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.petService.ExecuteCommand(petID, req.Command, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result, "message": "Command executed successfully"})
}

// RestPet 让宠物休息
func (h *PetHandler) RestPet(c *gin.Context) {
	petID := c.Param("id")
	
	err := h.petService.RestPet(petID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pet is now resting"})
}

// FeedPet 给宠物喂食
func (h *PetHandler) FeedPet(c *gin.Context) {
	petID := c.Param("id")
	
	type FeedRequest struct {
		Amount int `json:"amount"`
	}
	
	var req FeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Amount = 20 // 默认喂食量
	}

	err := h.petService.FeedPet(petID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pet has been fed"})
}

// SocializePet 让宠物社交
func (h *PetHandler) SocializePet(c *gin.Context) {
	petID := c.Param("id")
	
	err := h.petService.SocializePet(petID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pet is now socializing"})
}

// GetPetStatus 获取宠物详细状态
func (h *PetHandler) GetPetStatus(c *gin.Context) {
	petID := c.Param("id")
	
	status, err := h.petService.GetPetStatus(petID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetPetFriends 获取宠物朋友列表
func (h *PetHandler) GetPetFriends(c *gin.Context) {
	petID := c.Param("id")
	
	pet, exists := h.petService.GetPet(petID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pet_id":  petID,
		"pet_name": pet.Name,
		"friends": pet.Friends,
		"count":   len(pet.Friends),
	})
}