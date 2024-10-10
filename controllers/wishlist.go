package api

import (
	repository "wishlist-backend/repositories"

	"github.com/gin-gonic/gin"
)

type WishlistController struct {
	*AbstractController[repository.Wishlist, string]
	repo *repository.WishlistRepository
}

func (a *api) NewWishlistController() *WishlistController {
	return &WishlistController{
		repo: a.wishlistRepo,
		AbstractController: &AbstractController[repository.Wishlist, string]{
			api:          a,
			abstractRepo: a.wishlistRepo,
			paramToId:    func(s string) (string, error) { return s, nil },
			empty:        repository.Wishlist{},
		},
	}
}

func (controller *WishlistController) Init(router *gin.RouterGroup) {
	// router.GET("/admin", controller.GetAll)
	router.GET("/:id", controller.GetById)
	router.GET("/:id/items", controller.GetItems)
	router.POST("", controller.Add)
	router.PUT("/:id", controller.Update)
	router.PUT("", controller.Update)
	router.GET("", controller.GetAccessibleWishlists)
	router.POST("/:id/permission", controller.RegisterPermission)
	router.POST("/:id/permission/:password", controller.RegisterPermission)
}

func (controller *WishlistController) Add(c *gin.Context) {
	key := controller.GetAuthorization(c)
	if len(key) <= 0 {
		c.String(400, "You don't have a session key associated with your browser")
		return
	}

	model := repository.WishlistBody{}
	if err := c.BindJSON(&model); err != nil {
		c.String(401, "Invalid body was provided.")
		return
	}

	wishlist := repository.Wishlist{
		Model: repository.Model[string]{
			Id: model.Id,
		},
		Name:      model.Name,
		Ownership: key,
	}
	result, err := controller.abstractRepo.Add(wishlist)

	if err != nil {
		c.String(400, err.Error())
		return
	}

	c.IndentedJSON(201, repository.UnlockedWishlist{
		Model:     result.Model,
		Name:      result.Name,
		Password:  result.Password,
		Ownership: result.Ownership,
	})
}

func (controller *WishlistController) GetItems(c *gin.Context) {
	id, err := controller.ValidateId(c.Param("id"))

	if err != nil {
		c.String(401, "An invalid ID was provided.")
		return
	}

	items, err := controller.repo.GetItems(*id)

	if err != nil {
		c.String(500, "Something went wrong on the server.")
		c.Error(err)
		return
	}

	c.IndentedJSON(200, items)
}

func (controller *WishlistController) GetAccessibleWishlists(c *gin.Context) {
	key := controller.GetAuthorization(c)
	if len(key) <= 0 {
		c.String(400, "You don't have a session key associated with your browser")
		return
	}

	wls, err := controller.repo.GetOwnedWishlists(key)
	if err != nil {
		c.String(400, "Something went wrong")
		return
	}

	wishlists, err := controller.repo.GetSavedWishlists(key)
	if err != nil {
		c.String(400, "Something went wrong")
		return
	}

	for _, wl := range wls {
		wishlists = append(wishlists, repository.WishlistPermissioned{
			Wishlist:    wl,
			Permissions: "EDIT",
		})
	}

	c.IndentedJSON(200, wishlists)
}

func (controller *WishlistController) RegisterPermission(c *gin.Context) {
	key := controller.GetAuthorization(c)
	if len(key) <= 0 {
		c.String(400, "You don't have a session key associated with your browser")
		return
	}

	err := controller.repo.RegisterPermission(c.Param("id"), key, c.Param("password"))

	if err != nil {
		c.String(500, "Something went wrong.")
		return
	}

	c.String(200, "OK")
}
