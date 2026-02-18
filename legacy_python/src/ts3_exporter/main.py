import time
import logging
import ts3
from prometheus_client import start_http_server
from concurrent.futures import ThreadPoolExecutor

from . import config
from . import metrics

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class Teamspeak3MetricService:
    def __init__(self, server_config):
        self.server_config = server_config
        self.host = server_config.host
        self.port = server_config.port
        self.username = server_config.username
        self.password = server_config.password
        self.server_name = server_config.name
        self.serverQueryService = None

    def connect(self):
        try:
            # Connect to TS3 Server Query
            self.serverQueryService = ts3.TS3Server(self.host, self.port)
            self.serverQueryService.login(self.username, self.password)
            return True
        except Exception as e:
            logger.error(f"Failed to connect to {self.server_name} ({self.host}:{self.port}): {e}")
            return False

    def disconnect(self):
        if self.serverQueryService:
            try:
                self.serverQueryService.disconnect()
            except Exception as e:
                logger.warning(f"Error disconnecting from {self.server_name}: {e}")
            finally:
                self.serverQueryService = None

    def collect(self):
        if not self.connect():
            return

        try:
            serverlistResponse = self.serverQueryService.serverlist()
            # Check if command was successful. python-ts3 raises TS3Error on failure usually,
            # but the original code checked response['msg'].
            # Let's trust the original logic but wrap in try-except for safety.

            # The library returns a TS3QueryResponse object.
            # data is a list of dictionaries.

            servers = serverlistResponse.data

            for server in servers:
                virtualserver_id = server.get('virtualserver_id')
                if not virtualserver_id:
                    continue

                try:
                    self.serverQueryService.use(virtualserver_id)
                    serverinfoResponse = self.serverQueryService.send_command('serverinfo')
                    serverinfo = serverinfoResponse.data[0]
                    virtualserver_name = serverinfo.get('virtualserver_name', 'Unknown')

                    # Update Server Metrics
                    for metric_name in metrics.METRICS_NAMES:
                        if metric_name in serverinfo:
                            try:
                                value = float(serverinfo[metric_name])
                                metrics.PROMETHEUS_METRICS[metric_name].labels(
                                    server_name=self.server_name,
                                    virtualserver_name=virtualserver_name
                                ).set(value)
                            except (ValueError, TypeError):
                                pass

                    # Update Player Metrics
                    # Get detailed client list using options
                    clientlistResponse = self.serverQueryService.send_command(
                        'clientlist', 
                        opts=['uid', 'away', 'voice', 'times', 'groups', 'info', 'country', 'ip', 'badges']
                    )

                    players = clientlistResponse.data
                    for player in players:
                        client_nickname = player.get('client_nickname', '')
                        if client_nickname == "serveradmin":
                            continue

                        # Extract labels
                        # We use .get() for all fields to avoid KeyErrors
                        metrics.PLAYER_ONLINE.labels(
                            server_name=self.server_name,
                            virtualserver_name=virtualserver_name,
                            player_id=player.get('player_id', ''),
                            nickname=player.get('nickname', ''),
                            clid=player.get('clid', ''),
                            cid=player.get('cid', ''),
                            client_database_id=player.get('client_database_id', ''),
                            client_nickname=client_nickname,
                            client_type=player.get('client_type', ''),
                            client_away=player.get('client_away', ''),
                            client_away_message=player.get('client_away_message', ''),
                            client_flag_talking=player.get('client_flag_talking', ''),
                            client_input_muted=player.get('client_input_muted', ''),
                            client_output_muted=player.get('client_output_muted', ''),
                            client_input_hardware=player.get('client_input_hardware', ''),
                            client_output_hardware=player.get('client_output_hardware', ''),
                            client_talk_power=player.get('client_talk_power', ''),
                            client_is_talker=player.get('client_is_talker', ''),
                            client_is_priority_speaker=player.get('client_is_priority_speaker', ''),
                            client_is_recording=player.get('client_is_recording', ''),
                            client_is_channel_commander=player.get('client_is_channel_commander', ''),
                            client_unique_identifier=player.get('client_unique_identifier', ''),
                            client_servergroups=player.get('client_servergroups', ''),
                            client_channel_group_id=player.get('client_channel_group_id', ''),
                            client_channel_group_inherited_channel_id=player.get('client_channel_group_inherited_channel_id', ''),
                            client_version=player.get('client_version', ''),
                            client_platform=player.get('client_platform', ''),
                            client_idle_time=player.get('client_idle_time', ''),
                            client_created=player.get('client_created', ''),
                            client_lastconnected=player.get('client_lastconnected', ''),
                            client_country=player.get('client_country', ''),
                            connection_client_ip=player.get('connection_client_ip', ''),
                            client_badges=player.get('client_badges', '')
                        ).set(1)

                except Exception as e:
                    logger.error(f"Error processing virtual server {virtualserver_id} on {self.server_name}: {e}")

        except Exception as e:
            logger.error(f"Error collecting metrics from {self.server_name}: {e}")
        finally:
            self.disconnect()

def process_server(server_config):
    service = Teamspeak3MetricService(server_config)
    service.collect()

# Wrapper to avoid pickling issues if any, though not expected with ThreadPoolExecutor
def process_server_wrapper(server_config):
    try:
        process_server(server_config)
    except Exception as e:
        logger.error(f"Unhandled exception for server {server_config.name}: {e}")

class TeamSpeakExporter:
    def __init__(self):
        self.conf = config.load_config()

    def run(self):
        # Start Prometheus HTTP server
        logger.info(f"Starting metrics server on port {self.conf.metrics_port}")
        start_http_server(self.conf.metrics_port)

        logger.info(f"Monitoring {len(self.conf.servers)} servers with interval {self.conf.read_interval}s")

        while True:
            start_time = time.time()

            # Run collection for all servers in parallel
            with ThreadPoolExecutor(max_workers=max(1, len(self.conf.servers))) as executor:
                executor.map(process_server_wrapper, self.conf.servers)

            elapsed = time.time() - start_time
            sleep_time = max(0, self.conf.read_interval - elapsed)
            if sleep_time > 0:
                time.sleep(sleep_time)

def main():
    exporter = TeamSpeakExporter()
    exporter.run()

if __name__ == '__main__':
    main()
