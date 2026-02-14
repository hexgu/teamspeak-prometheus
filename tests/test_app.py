import unittest
from unittest.mock import MagicMock, patch
import sys
import os

# Add src directory to path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../src')))

from ts3_exporter import main as app
from ts3_exporter import config
from ts3_exporter import metrics

class TestTeamspeak3MetricService(unittest.TestCase):

    def setUp(self):
        # Reset metrics? Gauges are global.
        # It's better to verify calls on the gauge object if we can, but since we modify global state...
        # We can just ignore previous state or check for existence.
        pass

    @patch('ts3_exporter.main.ts3.TS3Server')
    def test_collect_metrics(self, mock_ts3_server):
        # Setup mock
        mock_server_instance = MagicMock()
        mock_ts3_server.return_value = mock_server_instance

        # Mock login success
        mock_server_instance.login.return_value = True

        # Mock serverlist
        mock_serverlist_resp = MagicMock()
        mock_serverlist_resp.data = [{'virtualserver_id': '1'}]
        mock_server_instance.serverlist.return_value = mock_serverlist_resp

        # Mock send_command for serverinfo and clientlist
        def mock_send_command(cmd, keys=None, opts=None, args=None):
            resp = MagicMock()
            if cmd == 'serverinfo':
                resp.data = [{
                    'virtualserver_name': 'Test Server',
                    'virtualserver_clientsonline': '10',
                    'virtualserver_uptime': '1000'
                }]
            elif cmd == 'clientlist':
                resp.data = [{
                    'client_nickname': 'User1',
                    'nickname': 'User1',
                    'player_id': '123',
                    'client_type': '0'
                }]
            else:
                resp.data = []
            return resp
        
        mock_server_instance.send_command.side_effect = mock_send_command

        # Create config
        server_conf = config.ServerConfig(
            name="My TS3",
            host="localhost",
            port=10011,
            username="admin",
            password="password"
        )

        # Create service
        service = app.Teamspeak3MetricService(server_conf)

        service.collect()

        # Verify connection
        mock_ts3_server.assert_called_with("localhost", 10011)
        mock_server_instance.login.assert_called_with("admin", "password")
        mock_server_instance.use.assert_called_with('1')

        # Verify server metrics
        # virtualserver_clientsonline
        metric = metrics.PROMETHEUS_METRICS['virtualserver_clientsonline']
        # We need to find the sample for our server
        found_server_metric = False
        for sample in metric.collect()[0].samples:
            if sample.labels['server_name'] == "My TS3" and sample.labels['virtualserver_name'] == "Test Server":
                self.assertEqual(sample.value, 10.0)
                found_server_metric = True
                break
        self.assertTrue(found_server_metric, "Server metric not found")

        # Verify player online metric
        player_metric = metrics.PLAYER_ONLINE

        found_player = False
        for sample in player_metric.collect()[0].samples:
            # Check labels
            labels = sample.labels
            if labels['server_name'] == "My TS3" and labels['client_nickname'] == "User1":
                self.assertEqual(sample.value, 1.0)
                found_player = True
                break

        self.assertTrue(found_player, "Player metric not found")

if __name__ == '__main__':
    unittest.main()
