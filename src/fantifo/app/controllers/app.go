package controllers

import (
	"github.com/revel/revel"
)

type Application struct {
	*revel.Controller
}

func (c Application) Index() revel.Result {
	return c.Render()
}

func (c Application) EnterDemo(position string) revel.Result {
	c.Validation.Required(position)

	if c.Validation.HasErrors() {
		c.Flash.Error("Please choose a position")
		return c.Redirect(Application.Index)
	}

	return c.Redirect("/fantifo?user=%s", position)
}
