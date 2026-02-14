from prometheus_client import Gauge

METRICS_PREFIX = 'teamspeak_'

METRICS_NAMES = [
    'connection_bandwidth_received_last_minute_total',
    'connection_bandwidth_received_last_second_total',
    'connection_bandwidth_sent_last_minute_total',
    'connection_bandwidth_sent_last_second_total',
    'connection_bytes_received_control',
    'connection_bytes_received_keepalive',
    'connection_bytes_received_speech',
    'connection_bytes_received_total',
    'connection_bytes_sent_control',
    'connection_bytes_sent_keepalive',
    'connection_bytes_sent_speech',
    'connection_bytes_sent_total',
    'connection_filetransfer_bandwidth_received',
    'connection_filetransfer_bandwidth_sent',
    'connection_filetransfer_bytes_received_total',
    'connection_filetransfer_bytes_sent_total',
    'connection_packets_received_control',
    'connection_packets_received_keepalive',
    'connection_packets_received_speech',
    'connection_packets_received_total',
    'connection_packets_sent_control',
    'connection_packets_sent_keepalive',
    'connection_packets_sent_speech',
    'connection_packets_sent_total',
    'virtualserver_channelsonline',
    'virtualserver_client_connections',
    'virtualserver_clientsonline',
    'virtualserver_maxclients',
    'virtualserver_month_bytes_downloaded',
    'virtualserver_month_bytes_uploaded',
    'virtualserver_query_client_connections',
    'virtualserver_queryclientsonline',
    'virtualserver_reserved_slots',
    'virtualserver_total_bytes_downloaded',
    'virtualserver_total_bytes_uploaded',
    'virtualserver_total_packetloss_control',
    'virtualserver_total_packetloss_keepalive',
    'virtualserver_total_packetloss_speech',
    'virtualserver_total_packetloss_total',
    'virtualserver_total_ping',
    'virtualserver_uptime'
]

# Create global metrics dictionary
PROMETHEUS_METRICS = {}

# Initialize metrics with 'server_name' label
for metric_name in METRICS_NAMES:
    PROMETHEUS_METRICS[metric_name] = Gauge(
        METRICS_PREFIX + metric_name,
        METRICS_PREFIX + metric_name,
        ['server_name', 'virtualserver_name']
    )

# Initialize player_online metric
PLAYER_ONLINE = Gauge('teamspeak_player_online', 'Online players',
    [
        'server_name',
        'virtualserver_name',
        'player_id',
        'nickname',
        'clid',
        'cid',
        'client_database_id',
        'client_nickname',
        'client_type',
        'client_away',
        'client_away_message',
        'client_flag_talking',
        'client_input_muted',
        'client_output_muted',
        'client_input_hardware',
        'client_output_hardware',
        'client_talk_power',
        'client_is_talker',
        'client_is_priority_speaker',
        'client_is_recording',
        'client_is_channel_commander',
        'client_unique_identifier',
        'client_servergroups',
        'client_channel_group_id',
        'client_channel_group_inherited_channel_id',
        'client_version',
        'client_platform',
        'client_idle_time',
        'client_created',
        'client_lastconnected',
        'client_country',
        'connection_client_ip',
        'client_badges',
    ]
)
