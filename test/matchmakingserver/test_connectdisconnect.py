#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class ConnectDisconnectTest(common.TestCase):
    def test_connect_disconnect(self):
        """c1 connects to the matchmaking server then disconnects. The game
        doesn't start.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 1)

        # Disconnect the client by killing it.
        # The sleeps here ensure the MS server and client are given enough time
        # to interact for log entries to be written.
        time.sleep(5)
        clients[0].kill()
        time.sleep(5)

        starting_game_found = False
        player_count_found = False
        with open(ms_srv.local_log_path) as log_file:
            for line in log_file:
                if "Starting Game" in line:
                    starting_game_found = True
                    continue
                elif "1 players" in line:
                    player_count_found = True
                    continue

        self.assertTrue(player_count_found,
                        "MS server should have been connected to")
        self.assertFalse(starting_game_found, "Game should not have started")

if __name__ == "__main__":
    unittest.main()
