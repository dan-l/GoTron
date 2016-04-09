#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class TooLittlePlayersTest(unittest.TestCase):
    def test_too_little_players(self):
        """c1 connects to the matchmaking server. Game doesn't start and we see
        one player waiting.
        """
        ms_srv_port = 2222
        ms_srv = common.MatchMakingServer(ms_srv_port)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv_port, 1)

        print "Sleeping for more than the MS server timeout"
        time.sleep(14)

        starting_game_found = False
        player_count_found = False
        with open(os.path.join(common.MATCHMAKING_DIR,
                               ms_srv.local_log_filename)) as log_file:
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

        ms_srv.kill()
        ms_srv.wait()
        for client in clients:
            client.kill()
            client.wait()

if __name__ == "__main__":
    unittest.main()
    print "Please issue Ctrl-C to try and kill any zombie processes."
