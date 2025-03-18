package prometheus

import "github.com/prometheus/client_golang/prometheus"

var (
	wsConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_active_connections",
		Help: "Количество активных websocket соединений",
	})

	messageSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "websocket_message_sent_total",
		Help: "Общее количество отправленных сообщений через websocket",
	})
)

func init() {
	prometheus.MustRegister(wsConnections)
	prometheus.MustRegister(messageSent)
}
