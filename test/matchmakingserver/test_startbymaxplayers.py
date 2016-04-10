#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class StartByMaxPlayersTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_start_by_max_players(self):
        """c1,..., c6 connect to the matchmaking server. The matchmaking server
        detects that the room size limit is reached and the game starts.
        """
        ms_srv_port = 2222
        ms_srv = common.MatchMakingServer(ms_srv_port)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv_port, 6)

        # Wait for a while to make sure MS server logs that one client is
        # connected.
        time.sleep(7)

        starting_game_found = False
        one_player_found = False
        two_players_found = False
        three_players_found = False
        four_players_found = False
        five_players_found = False
        six_players_found = False
        with open(os.path.join(common.MATCHMAKING_DIR,
                               ms_srv.local_log_filename)) as log_file:
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
                elif "3 players" in line:
                    three_players_found = True
                    continue
                elif "4 players" in line:
                    four_players_found = True
                    continue
                elif "5 players" in line:
                    five_players_found = True
                    continue
                elif "6 players" in line:
                    six_players_found = True
                    continue

        self.assertTrue(one_player_found,
                        ("MS server should have been connected to by one "
                         "player at some point"))
        self.assertTrue(two_players_found,
                        ("MS server should have been connected to by two "
                         "players at some point"))
        self.assertTrue(three_players_found,
                        ("MS server should have been connected to by three "
                         "players at some point"))
        self.assertTrue(four_players_found,
                        ("MS server should have been connected to by four "
                         "players at some point"))
        self.assertTrue(five_players_found,
                        ("MS server should have been connected to by five "
                         "players at some point"))
        self.assertTrue(six_players_found,
                        ("MS server should have been connected to by six "
                         "players at some point"))
        self.assertTrue(starting_game_found, "Game should have started")

        ms_srv.kill()
        ms_srv.wait()
        for client in clients:
            client.kill()
            client.wait()

if __name__ == "__main__":
    unittest.main()
