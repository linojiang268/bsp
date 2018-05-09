package repos

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-ozzo/ozzo-dbx"
	"time"
	"xungewang.cn/bsp/apis"
	"xungewang.cn/bsp/app"
	"xungewang.cn/bsp/models"
)

type (
	PositionRepo interface {
		FindStations(ctx app.RequestScope, signals map[string]apis.Signal) ([]models.Station, error)
		RecordUnknownSignals(ctx app.RequestScope, signals []apis.Signal)
	}

	defaultPositionRepo struct{}
)

func (repo *defaultPositionRepo) FindStations(ctx app.RequestScope, signals map[string]apis.Signal) ([]models.Station, error) {
	// we collect station ids, i.e., keys of signals map
	// Note: dbx.In() takes ...interface{} as the second argument, so we have to convert ids as []interface.
	ids := make([]interface{}, 0, len(signals))
	for id := range signals {
		ids = append(ids, id)
	}

	// query against database to find stations
	var stations []models.Station // note that dbx reject slice of pointers to structure as its argument
	err := ctx.Db().Select("id", "lat", "lng").From("base_stations").
		Where(dbx.In("id", ids...)).Limit(int64(len(ids))).All(&stations)

	return stations, err
}

func (repo *defaultPositionRepo) RecordUnknownSignals(ctx app.RequestScope, signals []apis.Signal) {
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	for _, signal := range signals {
		if _, err := ctx.Db().Insert("unknown_signals", dbx.Params{
			"mnc":        signal.Mnc,
			"lac":        signal.Lac,
			"cid":        signal.Cid,
			"created_at": createdAt,
		}).Execute(); err != nil {
			log.Errorf("failed to persist unknown signal(%s): %s", signal, err)
		}
	}
}

// NewPositionRepo create instance of PositionRepo
func NewPositionRepo() *defaultPositionRepo {
	return &defaultPositionRepo{}
}
