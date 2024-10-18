package controllers

import (
	"fmt"
	"newsletter/tools"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

type dateRequested struct {
	Month string
	Year  string
}

func SubsPerMonth(c echo.Context, app *pocketbase.PocketBase) error {
	year, err := strconv.ParseInt(c.PathParam("year"), 10, 64)
	if err != nil {
		return err
	}
	month, err := strconv.ParseInt(c.PathParam("month"), 10, 64)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(created) as subscriptions, STRFTIME("Dia %%d", created) as day FROM
		(
			SELECT created FROM subscribers 
			WHERE strftime("%%Y-%%m", created) = '%d-%d'
		) 
		as gru GROUP BY STRFTIME("%%d", gru.created);
	`, year, month)

	var res []struct {
		Subscriptions int
		Day           string
	}

	app.Dao().DB().NewQuery(query).All(&res)

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{}))

	days := make([]string, len(res))
	subsCount := make([]opts.BarData, len(res))
	for i, v := range res {
		days[i] = v.Day
		subsCount[i] = opts.BarData{Value: v.Subscriptions}
	}

	xAxis := bar.SetXAxis(days)
	xAxis.AddSeries("Numbero de subscritores", subsCount)

	bar.Render(c.Response().Writer)
	return nil
}

func SubsStatusCount(c echo.Context, app *pocketbase.PocketBase) error {
	var res struct {
		UnVerifyCount    int
		SubscribersCount int
		UnSubscribeCount int
		Total            int
	}

	query := fmt.Sprintf(`
		SELECT 
			(SELECT COUNT(id) FROM subscribers WHERE status = "%s") as UnVerifyCount, 
			(SELECT COUNT(id) FROM subscribers WHERE status = "%s") as SubscribersCount, 
			(SELECT COUNT(id) FROM subscribers WHERE status = "%s") as UnSubscribeCount, 
			(SELECT COUNT(id) FROM subscribers) as Total 
		FROM subscribers LIMIT 1 
	`, tools.SubscriberStatus[tools.SUBSCRIBER_UNVERIFIED_STATUS], tools.SubscriberStatus[tools.SUBSCRIBER_VERIFIED_STATUS], tools.SubscriberStatus[tools.SUBSCRIBER_UNSUBSCRIBE_STATUS])
	app.Dao().DB().NewQuery(query).Row(&res.UnVerifyCount, &res.SubscribersCount, &res.UnSubscribeCount, &res.Total)

	pie := charts.NewPie()
	pie.AddSeries("pie", []opts.PieData{
		{Value: res.UnSubscribeCount, Name: "Desuscritos"},
		{Value: res.SubscribersCount, Name: "Suscritos"},
		{Value: res.UnVerifyCount, Name: "No verificados"},
	}).SetSeriesOptions(charts.WithLabelOpts(
		opts.Label{
			Show:      true,
			Formatter: "{b}: {c}",
		},
	))

	pie.Render(c.Response().Writer)
	return nil
}
