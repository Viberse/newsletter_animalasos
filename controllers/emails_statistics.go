package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

type SesStadisticsStruct struct {
	Bounces          int64 `json:"bounces"`
	Complaints       int64 `json:"complaints"`
	DeliveryAttempts int64 `json:"deliveryAttempts"`
	Rejects          int64 `json:"rejects"`
}

func generateClicksLine(year int64, month int64, app *pocketbase.PocketBase) *charts.Line {
	query := fmt.Sprintf(`
		SELECT 
			SUM(counter) as subscriptions, STRFTIME("%%d", created) as day 
		FROM 
			clicks 
		WHERE strftime("%%Y-%%m", created) = '%d-%d' 
		GROUP BY STRFTIME("%%d", created)
	`, year, month)

	var res []struct {
		Subscriptions int
		Day           string
	}

	app.Dao().DB().NewQuery(query).All(&res)

	line := charts.NewLine()
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{}))

	days := make([]int, len(res))
	subsCount := make([]opts.LineData, len(res))
	for i, v := range res {
		day, _ := strconv.ParseInt(v.Day, 10, 64)
		days[i] = int(day)
		subsCount[i] = opts.LineData{Value: v.Subscriptions}
	}

	xAxis := line.SetXAxis(days)
	xAxis.
		AddSeries("Numbero de clicks en un dia", subsCount).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				ShowSymbol: true,
			}),
			charts.WithLabelOpts(opts.Label{
				Show: true,
			}),
		)

	return line
}

func SesStadistics(c echo.Context, app *pocketbase.PocketBase) error {
	year, err := strconv.ParseInt(c.PathParam("year"), 10, 64)
	if err != nil {
		return err
	}
	month, err := strconv.ParseInt(c.PathParam("month"), 10, 64)
	if err != nil {
		return err
	}

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, "Invalid body")
	}

	res, err := sesCli.GetSendStatistics(c.Request().Context(), &ses.GetSendStatisticsInput{})
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, "Invalid body")
	}

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{}))
	xAxisData := make(map[int][]SesStadisticsStruct)
	xAxisLegend := make([]int, 0)
	for _, dataPoints := range res.SendDataPoints {
		timestamp := dataPoints.Timestamp
		if timestamp.Year() == int(year) && timestamp.Month() == time.Month(month) {
			if _, ok := xAxisData[timestamp.Day()]; ok {
				xAxisData[timestamp.Day()] = append(xAxisData[timestamp.Day()], SesStadisticsStruct{
					Bounces:          dataPoints.Bounces,
					Complaints:       dataPoints.Complaints,
					DeliveryAttempts: dataPoints.DeliveryAttempts,
					Rejects:          dataPoints.Rejects,
				})
			} else {
				xAxisData[timestamp.Day()] = []SesStadisticsStruct{{
					Bounces:          dataPoints.Bounces,
					Complaints:       dataPoints.Complaints,
					DeliveryAttempts: dataPoints.DeliveryAttempts,
					Rejects:          dataPoints.Rejects,
				}}
				xAxisLegend = append(xAxisLegend, timestamp.Day())
			}
		}
	}

	xAxis := bar.SetXAxis(xAxisLegend)

	complaints := make([]opts.BarData, len(xAxisLegend))
	deliveryAttempts := make([]opts.BarData, len(xAxisLegend))
	bounces := make([]opts.BarData, len(xAxisLegend))
	rejects := make([]opts.BarData, len(xAxisLegend))
	sort.Ints(xAxisLegend)
	for i := 0; i < len(xAxisLegend); i++ {
		day := xAxisLegend[i]
		complaintsCount := 0
		deliveryAttemptsCount := 0
		bouncesCount := 0
		rejectsCount := 0
		for e := 0; e < len(xAxisData[day]); e++ {
			complaintsCount += int(xAxisData[day][e].Complaints)
			deliveryAttemptsCount += int(xAxisData[day][e].DeliveryAttempts)
			bouncesCount += int(xAxisData[day][e].Bounces)
			rejectsCount += int(xAxisData[day][e].Rejects)
		}
		complaints[i] = opts.BarData{Value: complaintsCount}
		deliveryAttempts[i] = opts.BarData{Value: deliveryAttemptsCount}
		bounces[i] = opts.BarData{Value: bouncesCount}
		rejects[i] = opts.BarData{Value: rejectsCount}
	}
	xAxis.AddSeries("Complaints", complaints).AddSeries("Delivery attempts", deliveryAttempts).AddSeries("Bounces", bounces).AddSeries("Rejects", rejects)
	xAxis.SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show:     true,
			Position: "top",
		}),
	)

	line := generateClicksLine(year, month, app)
	if len(bar.XAxisList) > len(line.XAxisList) {
		bar.Overlap(line)
		bar.Render(c.Response().Writer)
		return nil
	}
	line.Overlap(bar)
	line.Render(c.Response().Writer)

	return nil
}
