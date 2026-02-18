package collector

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hexgu/teamspeak-prometheus/internal/config"
	"github.com/hexgu/teamspeak-prometheus/internal/ts3"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serverMetrics = []string{
		"connection_bandwidth_received_last_minute_total",
		"connection_bandwidth_received_last_second_total",
		"connection_bandwidth_sent_last_minute_total",
		"connection_bandwidth_sent_last_second_total",
		"connection_bytes_received_control",
		"connection_bytes_received_keepalive",
		"connection_bytes_received_speech",
		"connection_bytes_received_total",
		"connection_bytes_sent_control",
		"connection_bytes_sent_keepalive",
		"connection_bytes_sent_speech",
		"connection_bytes_sent_total",
		"connection_filetransfer_bandwidth_received",
		"connection_filetransfer_bandwidth_sent",
		"connection_filetransfer_bytes_received_total",
		"connection_filetransfer_bytes_sent_total",
		"connection_packets_received_control",
		"connection_packets_received_keepalive",
		"connection_packets_received_speech",
		"connection_packets_received_total",
		"connection_packets_sent_control",
		"connection_packets_sent_keepalive",
		"connection_packets_sent_speech",
		"connection_packets_sent_total",
		"virtualserver_channelsonline",
		"virtualserver_client_connections",
		"virtualserver_clientsonline",
		"virtualserver_maxclients",
		"virtualserver_month_bytes_downloaded",
		"virtualserver_month_bytes_uploaded",
		"virtualserver_query_client_connections",
		"virtualserver_queryclientsonline",
		"virtualserver_reserved_slots",
		"virtualserver_total_bytes_downloaded",
		"virtualserver_total_bytes_uploaded",
		"virtualserver_total_packetloss_control",
		"virtualserver_total_packetloss_keepalive",
		"virtualserver_total_packetloss_speech",
		"virtualserver_total_packetloss_total",
		"virtualserver_total_ping",
		"virtualserver_uptime",
	}
)

type TS3Collector struct {
	config *config.Config
	pools  map[string]*ts3.Pool
	descs  map[string]*prometheus.Desc

	playerOnlineDesc *prometheus.Desc
}

func NewTS3Collector(cfg *config.Config) *TS3Collector {
	descs := make(map[string]*prometheus.Desc)
	for _, name := range serverMetrics {
		descs[name] = prometheus.NewDesc(
			"teamspeak_"+name,
			"TeamSpeak 3 Server Metric "+name,
			[]string{"server_name", "virtualserver_name"},
			nil,
		)
	}

	playerDesc := prometheus.NewDesc(
		"teamspeak_player_online",
		"Online players",
		[]string{
			"server_name", "virtualserver_name", "player_id", "nickname", "clid", "cid",
			"client_database_id", "client_nickname", "client_type", "client_away",
			"client_away_message", "client_flag_talking", "client_input_muted",
			"client_output_muted", "client_input_hardware", "client_output_hardware",
			"client_talk_power", "client_is_talker", "client_is_priority_speaker",
			"client_is_recording", "client_is_channel_commander",
			"client_unique_identifier", "client_servergroups", "client_channel_group_id",
			"client_channel_group_inherited_channel_id", "client_version", "client_platform",
			"client_idle_time", "client_created", "client_lastconnected",
			"client_country", "connection_client_ip", "client_badges",
		},
		nil,
	)

	c := &TS3Collector{
		config:           cfg,
		pools:            make(map[string]*ts3.Pool),
		descs:            descs,
		playerOnlineDesc: playerDesc,
	}

	for _, server := range cfg.Servers {
		s := server
		factory := func() (*ts3.Client, error) {
			return ts3.NewClient(s.Host, s.Port, s.Username, s.Password, s.Protocol, 10*time.Second)
		}
		key := s.Name
		if key == "" {
			key = fmt.Sprintf("%s:%d", s.Host, s.Port)
		}
		c.pools[key] = ts3.NewPool(factory, 5) // Max 5 connections per server
	}

	return c
}

func (c *TS3Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
	ch <- c.playerOnlineDesc
}

func (c *TS3Collector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, server := range c.config.Servers {
		wg.Add(1)
		go func(s config.ServerConfig) {
			defer wg.Done()
			c.collectServer(s, ch)
		}(server)
	}
	wg.Wait()
}

func (c *TS3Collector) collectServer(serverConf config.ServerConfig, ch chan<- prometheus.Metric) {
	key := serverConf.Name
	if key == "" {
		key = fmt.Sprintf("%s:%d", serverConf.Host, serverConf.Port)
	}
	pool := c.pools[key]

	// Get connection for serverlist
	client, err := pool.Get()
	if err != nil {
		log.Printf("Failed to get connection for %s: %v", serverConf.Name, err)
		return
	}

	resp, err := client.Execute("serverlist")
	if err != nil {
		log.Printf("Failed to list servers on %s: %v", serverConf.Name, err)
		pool.Discard(client)
		return
	}
	pool.Put(client)

	vServers := ts3.ParseResponse(resp)
	var wgVServer sync.WaitGroup

	for _, vServer := range vServers {
		vidStr := vServer["virtualserver_id"]
		if vidStr == "" {
			continue
		}

		wgVServer.Add(1)
		go func(vid string) {
			defer wgVServer.Done()
			c.collectVirtualServer(serverConf, pool, vid, ch)
		}(vidStr)
	}
	wgVServer.Wait()
}

func (c *TS3Collector) collectVirtualServer(serverConf config.ServerConfig, pool *ts3.Pool, vid string, ch chan<- prometheus.Metric) {
	client, err := pool.Get()
	if err != nil {
		log.Printf("Failed to get connection for vserver %s on %s: %v", vid, serverConf.Name, err)
		return
	}

	broken := false
	defer func() {
		if broken {
			pool.Discard(client)
		} else {
			pool.Put(client)
		}
	}()

	if _, err := client.Execute("use " + vid); err != nil {
		log.Printf("Failed to select vserver %s on %s: %v", vid, serverConf.Name, err)
		broken = true
		return
	}

	// 1. Server Info
	respInfo, err := client.Execute("serverinfo")
	if err != nil {
		log.Printf("Failed to get serverinfo for vserver %s on %s: %v", vid, serverConf.Name, err)
		broken = true
		return
	}

	infoMaps := ts3.ParseResponse(respInfo)
	if len(infoMaps) == 0 {
		return
	}
	info := infoMaps[0]
	vServerName := info["virtualserver_name"]

	for metricName, desc := range c.descs {
		if valStr, ok := info[metricName]; ok {
			if val, err := strconv.ParseFloat(valStr, 64); err == nil {
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					val,
					serverConf.Name,
					vServerName,
				)
			}
		}
	}

	// 2. Client List
	// opts: uid, away, voice, times, groups, info, country, ip, badges
	respClients, err := client.Execute("clientlist -uid -away -voice -times -groups -info -country -ip -badges")
	if err != nil {
		log.Printf("Failed to get clientlist for vserver %s on %s: %v", vid, serverConf.Name, err)
		broken = true
		return
	}

	clients := ts3.ParseResponse(respClients)
	for _, player := range clients {
		if player["client_nickname"] == "serveradmin" || player["client_type"] == "1" {
			continue // Skip query clients
		}

		// Extract labels
		labels := []string{
			serverConf.Name,
			vServerName,
			player["player_id"],
			player["nickname"], // This field might be missing or different key? standard is client_nickname
			player["clid"],
			player["cid"],
			player["client_database_id"],
			player["client_nickname"],
			player["client_type"],
			player["client_away"],
			player["client_away_message"],
			player["client_flag_talking"],
			player["client_input_muted"],
			player["client_output_muted"],
			player["client_input_hardware"],
			player["client_output_hardware"],
			player["client_talk_power"],
			player["client_is_talker"],
			player["client_is_priority_speaker"],
			player["client_is_recording"],
			player["client_is_channel_commander"],
			player["client_unique_identifier"],
			player["client_servergroups"],
			player["client_channel_group_id"],
			player["client_channel_group_inherited_channel_id"],
			player["client_version"],
			player["client_platform"],
			player["client_idle_time"],
			player["client_created"],
			player["client_lastconnected"],
			player["client_country"],
			player["connection_client_ip"],
			player["client_badges"],
		}

		ch <- prometheus.MustNewConstMetric(
			c.playerOnlineDesc,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)
	}
}
