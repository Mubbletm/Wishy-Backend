package api

import (
	"errors"
	"os"
	"reflect"
	"strings"
	repository "wishlist-backend/repositories"

	"github.com/gin-gonic/gin"
)

type Controller[M interface{}, I any] interface {
	Add(*gin.Context)
	GetById(*gin.Context)
	Update(*gin.Context)
	GetAll(*gin.Context)
}

// An abstract controller that fully implements the Controller[M, I] interface.
//
// Generic[M] is the type of the model of this controller.
//
// Generic[I] is the type of the model's ID of this controller. The ID of your model
// should always be named 'Id' in the database for it to work with this controller and its
// repository.
type AbstractController[M interface{}, I any] struct {
	Controller[M, I]
	api          *api
	abstractRepo repository.Repository[M, I]
	paramToId    func(string) (I, error)
	empty        M
}

func (controller *AbstractController[M, I]) GetById(c *gin.Context) {
	id, err := controller.ValidateId(c.Param("id"))

	if err != nil {
		c.String(401, err.Error())
		return
	}

	model, err := controller.abstractRepo.GetById(*id)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.String(404, "Not found")
			c.Error(os.ErrNotExist)
			return
		}
		c.Error(err)
		c.String(400, "Something went wrong")
		return
	}

	c.IndentedJSON(200, model)
}

func (controller *AbstractController[M, I]) GetAuthorization(c *gin.Context) string {
	return strings.TrimSpace(strings.Replace(strings.ToLower(c.GetHeader("Authorization")), "bearer", "", 1))
}

func (controller *AbstractController[M, I]) GetAll(c *gin.Context) {
	result, err := controller.abstractRepo.GetAll()

	if err != nil {
		c.String(400, "Something went wrong")
		c.Error(err)
		return
	}

	c.IndentedJSON(200, result)
}

func (controller *AbstractController[M, I]) Add(c *gin.Context) {
	model := controller.empty
	if err := c.BindJSON(&model); err != nil {
		c.String(401, "Invalid body was provided.")
		return
	}
	// controller.repo.RemoveId(&model)
	result, err := controller.abstractRepo.Add(model)

	if err != nil {
		c.String(400, err.Error())
		return
	}

	c.IndentedJSON(201, result)
}

func (controller *AbstractController[M, I]) Update(c *gin.Context) {
	id := c.Param("id")
	model := controller.empty

	if err := c.BindJSON(&model); err != nil {
		c.String(401, "Invalid body was provided.")
		return
	}

	searchId := reflect.ValueOf(model).FieldByName("Id").Interface().(I)

	if len(id) != 0 {
		var err error
		searchId, err = controller.paramToId(id)

		if err != nil {
			c.String(401, "id is not in the right format")
			return
		}
	}

	if err := controller.abstractRepo.Update(model, &searchId); err != nil {
		c.Error(err)
		c.String(400, "Something went wrong while updating the item in the database.")
		return
	}

	c.IndentedJSON(201, model)
}

func (controller *AbstractController[M, I]) ValidateId(id string) (*I, error) {
	if len(id) == 0 {
		return nil, os.ErrInvalid
	}

	convertedId, err := controller.paramToId(id)

	if err != nil {
		return nil, os.ErrInvalid
	}

	return &convertedId, nil
}
