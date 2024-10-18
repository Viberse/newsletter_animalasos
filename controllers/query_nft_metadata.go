package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

var _ models.Model = (*Nfts)(nil)

type Nfts struct {
	models.BaseModel

	Url string `db:"url" json:"url"`
}

func (m *Nfts) TableName() string {
	return "nfts"
}

type Attributes struct {
	Name  string
	Value string
}

func QueryNftMetadata(c echo.Context, app *pocketbase.PocketBase) error {
	params := c.QueryParams()
	page, err := strconv.Atoi(c.PathParam("pagina"))
	if err != nil {
		return c.String(400, "Invalid page")
	}

	db := app.Dao().DB()

	output := []*Nfts{}

	if len(params) != 0 {
		// animalsQuerys := []Attributes{}
		attributesExprs := map[string][]dbx.Expression{}
		// attributesExprs2 := []dbx.Expression{}

		for queryName, values := range params {
			for _, v := range values {
				// switch queryName {
				// case "ANIMAL":
				// 	animalsQuerys = append(animalsQuerys, Attributes{
				// 		Name:  "animal",
				// 		Value: v,
				// 	})
				// default:
				name := strings.Replace(strings.ToLower(queryName), " ", "_", -1)
				// attributesExprs2 = append(attributesExprs2,
				// 	dbx.HashExp{name: v},
				// )
				attributesExprs[name] = append(attributesExprs[name],
					dbx.HashExp{name: v},
				)
				// }
			}
		}

		query := app.Dao().
			ModelQuery(&Nfts{})

		attributesMerges := []dbx.Expression{}

		for _, v := range attributesExprs {
			attributesMerges = append(attributesMerges, dbx.Enclose(dbx.Or(v...)))
		}

		query.
			Where(dbx.And(attributesMerges...)).
			Limit(20).
			Offset(int64(page) * 20).
			All(&output)

		return c.JSON(200, output)
	}

	if err := db.NewQuery(fmt.Sprintf(`SELECT nfts.url FROM nfts LIMIT 20 OFFSET %v`, page*20)).All(&output); err != nil {
		println(err)
		return err
	}

	return c.JSON(200, output)
}
