#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class StartByTimeoutTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_start_by_timeout(self):
        """c1 connects to the matchmaking server. Then c2 connects to the
        matchmaking server. When the countdown timer expires, the game starts.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        time.sleep(2)

        _ = common.start_multiple_clients(ms_srv.port, 1)

        # Wait for a while to make sure MS server logs that one client is
        # connected.
        time.sleep(7)

        # Start client 2, then disconnect client 1 by killing client 1.
        client2 = common.Client(node_port=9003,
                                node_rpc_port=9002,
                                ms_port=ms_srv.port,
                                http_srv_port=9001)
        client2.start()
        time.sleep(10)

        starting_game_found = False
        one_player_found = False
        two_players_found = False
        with open(ms_srv.local_log_path) as log_file:
            for line in log_file:
                if "Starting Game" in line:
                    starting_game_found = True
                    continue
                elif "1 players" in line:
                    one_player_found = True
                    continue
                elif "2 players" in line:
                    two_players_found = True
                    continue

        self.assertTrue(one_player_found,
                        ("MS server should have been connected to by one "
                         "player at some point"))
        self.assertTrue(two_players_found,
                        ("MS server should have been connected to by two "
                         "players at some point"))
        self.assertTrue(starting_game_found, "Game should have started")

if __name__ == "__main__":
    unittest.main()
