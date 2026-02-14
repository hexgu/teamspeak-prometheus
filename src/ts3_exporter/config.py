import os
import yaml
import logging

logger = logging.getLogger(__name__)

class ServerConfig:
    def __init__(self, name, host, port, username, password):
        self.name = name
        self.host = host
        self.port = port
        self.username = username
        self.password = password

class Config:
    def __init__(self):
        self.servers = []
        self.metrics_port = 8000
        self.read_interval = 60

    def add_server(self, server):
        self.servers.append(server)

def load_config(config_path='config.yaml'):
    config = Config()

    if os.path.exists(config_path):
        logger.info(f"Loading configuration from {config_path}")
        with open(config_path, 'r') as f:
            data = yaml.safe_load(f)

        config.metrics_port = data.get('metrics_port', 8000)
        config.read_interval = data.get('read_interval', 60)

        servers_data = data.get('servers', [])
        for s in servers_data:
            server_config = ServerConfig(
                name=s.get('name', 'Unknown Server'),
                host=s['host'],
                port=s.get('port', 10011),
                username=s.get('username', 'serveradmin'),
                password=s.get('password', '')
            )
            config.add_server(server_config)
    else:
        logger.info("Config file not found. Falling back to environment variables.")
        # Fallback to environment variables
        host = os.environ.get('TEAMSPEAK_HOST')
        if host:
            server_config = ServerConfig(
                name=os.environ.get('TEAMSPEAK_SERVER_NAME', 'Default Server'),
                host=host,
                port=int(os.environ.get('TEAMSPEAK_PORT', 10011)),
                username=os.environ.get('TEAMSPEAK_USERNAME', 'serveradmin'),
                password=os.environ.get('TEAMSPEAK_PASSWORD', '')
            )
            config.add_server(server_config)

        config.metrics_port = int(os.environ.get('METRICS_PORT', 8000))
        # READ_INTERVAL_IN_SECONDS was hardcoded to 60 in app.py, keeping it as default here but allow env var if needed?
        # app.py didn't use env var for read interval, so sticking to default.

    if not config.servers:
        logger.warning("No TeamSpeak servers configured!")

    return config
