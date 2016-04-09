#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class DisconnectBeforeGameStartTest(unittest.TestCase):
    def test_connect_disconnect(self):
        """c1 connects to the matchmaking server. c2 connects to the matchmaking
        server, c1 disconnects before countdown timer expires. Game doesn't
        start and we see one player waiting.
        """
        ms_srv_port = 2222
        ms_srv = common.MatchMakingServer(ms_srv_port)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv_port, 1)

        # Wait for a while to make sure MS server logs that one client is
        # connected.
        time.sleep(7)

        # Start client 2, then disconnect client 1 by killing client 1.
        client2 = common.Client(node_port=9003,
                                node_rpc_port=9002,
                                ms_port=ms_srv_port,
                                http_srv_port=9001)
        client2.start()
        time.sleep(2)
        client1 = clients[0]
        client1.kill()
        time.sleep(5)

        starting_game_found = False
        one_player_found = False
        two_players_found = False
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

        self.assertTrue(one_player_found,
                        ("MS server should have been connected to by one "
                         "player at some point"))
        self.assertTrue(two_players_found,
                        ("MS server should have been connected to by two "
                         "players at some point"))
        self.assertFalse(starting_game_found, "Game should not have started")

        ms_srv.kill()
        ms_srv.wait()
        client2.kill()
        client2.wait()
        for client in clients:
            client.kill()
            client.wait()

if __name__ == "__main__":
    unittest.main()
    print "Please issue Ctrl-C to try and kill any zombie processes."
