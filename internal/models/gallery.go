package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GalleryState interface {
	List() []Painting
	Save(Painting) primitive.ObjectID

	One(primitive.ObjectID) Painting
	Delete(primitive.ObjectID) bool
	Update(primitive.ObjectID, bson.M) bool
}

type Gallery struct {
	state GalleryState
}

func NewGallery(state GalleryState) *Gallery {
	return &Gallery{state: state}
}

func (w *Gallery) AddProduct(p Painting) primitive.ObjectID {
	return w.state.Save(p)
}

func (w *Gallery) ListProducts() []Painting {
	return w.state.List()
}

func (w *Gallery) GetOnePainting(id primitive.ObjectID) Painting {
	return w.state.One(id)

}

func (w *Gallery) DeletePainting(id primitive.ObjectID) bool {
	return w.state.Delete(id)
}

func (w *Gallery) UpdatePainting(id primitive.ObjectID, update bson.M) bool {
	return w.state.Update(id, update)
}
