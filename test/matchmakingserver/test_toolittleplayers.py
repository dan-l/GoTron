#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class TooLittlePlayersTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_too_little_players(self):
        """c1 connects to the matchmaking server. Game doesn't start and we see
        one player waiting.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        time.sleep(2)

        _ = common.start_multiple_clients(ms_srv.port, 1)

        print "Sleeping for more than the MS server timeout"
        time.sleep(14)

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
