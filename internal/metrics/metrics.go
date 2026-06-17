package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	UsersCreatedTotal      = "users_new_total"
	APIUserRequestDuration = "api_user_request_duration_seconds"

	MethodTag   = "method"
	PostTagVal  = "post"
	GetTagVal   = "get"
)

var (
	UsersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: UsersCreatedTotal,
		Help: "Общее количество созданных пользователей с момента запуска",
	})

	APIUserRequest = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    APIUserRequestDuration,
		Help:    "Длительность обработки запросов к API пользователей",
		Buckets: prometheus.DefBuckets,
	}, []string{MethodTag})
)
