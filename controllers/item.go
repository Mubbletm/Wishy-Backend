package api

import (
	repository "wishlist-backend/repositories"
	"wishlist-backend/services/ogp"

	"github.com/gin-gonic/gin"
)

type ItemController struct {
	*AbstractController[repository.Item, string]
}

func (a *api) NewItemController() *ItemController {
	return &ItemController{
		&AbstractController[repository.Item, string]{
			api:          a,
			abstractRepo: a.itemRepo,
			paramToId:    func(s string) (string, error) { return s, nil },
			empty:        repository.Item{},
		},
	}
}

func (controller *ItemController) Init(router *gin.RouterGroup) {
	router.GET("", controller.GetAll)
	router.GET("/:id", controller.GetById)
	// router.GET("/:id/embed", controller.GetEmbed)
	router.PUT("/:id", controller.Update)
	router.PUT("", controller.Update)
	router.POST("/:wishlistId", controller.Add)
}

func (controller *ItemController) Add(c *gin.Context) {

	// DANGEROUS, IF WISHLIST ID TYPE CHANGES, THIS WILL BREAK
	id, err := controller.ValidateId(c.Param("wishlistId"))
	if err != nil {
		c.String(401, "An invalid ID was provided.")
		return
	}

	model := controller.empty
	if err := c.BindJSON(&model); err != nil {
		c.String(401, "Invalid body was provided.")
		return
	}

	data, err := ogp.GetOGPData(model.Url)
	if err != nil {
		c.String(401, "Invalid URL was provided.")
		return
	}

	model.WishlistId = *id
	model.Description = data.Description
	model.Image = data.Image
	model.Name = data.Title
	model.Url = data.Url

	// controller.repo.RemoveId(&model)
	result, err := controller.abstractRepo.Add(model)

	if err != nil {
		c.String(400, err.Error())
		return
	}

	c.IndentedJSON(201, result)
}
