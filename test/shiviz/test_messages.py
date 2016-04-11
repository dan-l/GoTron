#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class BasicTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_basic(self):
        """Play a game. All communication between nodes, and with the
        matchmaking server should be logged in a file.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 2)

        # Wait for a bit to make sure a game starts and is played for a while.
        time.sleep(common.MatchMakingServer.GAME_START_TIMEOUT * 1.1)

        found_join_msg = False
        found_start_game_msg = False
        with open(ms_srv.govector_log_path) as log_file:
            lines = log_file.readlines()
            for line in lines[1::2]:
                # Logged when receiving a Join RPC call from a node.
                if "AD: new node" in line:
                    found_join_msg = True
                # Logged when doing an RPC call to a node to start a game.
                elif "Rpc Call NodeService.StartGame" in line:
                    found_start_game_msg = True

        # TODO: Ensure that join and start game messages are found for all nodes,
        #       not just at least one.
        self.assertTrue(found_join_msg,
                        "MS should've logged that a node joined")
        self.assertTrue(found_start_game_msg,
                        "MS should've logged that it notified nodes of game start")

        for client in clients:
            found_join_msg = False
            found_start_game_msg = False
            found_send_interval_msg = False
            found_recv_interval_msg = False
            with open(client.govector_log_path) as log_file:
                lines = log_file.readlines()
                for line in lines[1::2]:
                    # Logged when doing a Join RPC call to MS.
                    if "Rpc Call Context.Join" in line:
                        found_join_msg = True
                    # Logged when receiving an RPC call to start the game.
                    elif "Rpc Called Start Game" in line:
                        found_start_game_msg = True
                    # Logged when sending an interval update message to peers.
                    # TODO: We need to test that we send interval updates to *ALL* peers.
                    elif "Sending: Interval update" in line:
                        found_send_interval_msg = True
                    # Logged when receiving an interval update msg from a peer.
                    # TODO: We need to test we get interval updates from all peers.
                    # TODO: We need to change the message to be specific to interval updates
                    elif "Received packet from" in line:
                        found_recv_interval_msg = True

            self.assertTrue(found_join_msg,
                            "Node should've logged that it contacted the MS server for joining")
            self.assertTrue(found_start_game_msg,
                            "Node should've logged that MS server said the game was starting")
            self.assertTrue(found_send_interval_msg,
                            "Node should've logged that it was sending an interval update")
            self.assertTrue(found_recv_interval_msg,
                            "Node should've logged that it received an interval update")

if __name__ == "__main__":
    unittest.main()
