package api

import (
	repository "wishlist-backend/repositories"

	"github.com/doug-martin/goqu/v9"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type api struct {
	httpClient   *gin.Engine
	itemRepo     *repository.ItemRepository
	wishlistRepo *repository.WishlistRepository
}

func New(db *goqu.Database) *api {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:4000"}

	httpClient := gin.Default()
	httpClient.Use(cors.New(config))

	apiObj := &api{
		httpClient:   httpClient,
		itemRepo:     repository.NewItemRepository(db),
		wishlistRepo: repository.NewWishlistRepository(db),
	}

	apiObj.NewItemController().Init(httpClient.Group("/item"))
	apiObj.NewWishlistController().Init(httpClient.Group("/wishlist"))

	return apiObj
}

func (a *api) Run(url string) error {
	return a.httpClient.Run(url)
}
