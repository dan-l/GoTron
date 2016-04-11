#!/usr/bin/env python2

import os
import re
import sys
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class BasicTest(common.TestCase):
    def test_basic(self):
        """Play a game. All communication between nodes, and with the
        matchmaking server should be logged in a file.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        common.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 2)

        # Wait for a bit to make sure a game starts and is played for a while.
        common.sleep(common.MatchMakingServer.GAME_START_TIMEOUT * 1.1)

        lines = []
        with open(ms_srv.govector_log_path) as log_file:
            lines = log_file.readlines()

        # Check that the MS interacts with all clients.
        for client in clients:
            join_msg = "AD: new node: IP: localhost:{}".format(client.node_port)
            found_join_msg = False
            start_game_msg = "Rpc Call NodeService.StartGame to localhost:{}".format(
                client.node_port)
            found_start_game_msg = False
            for line in lines[1::2]:
                # Logged when receiving a Join RPC call from a node.
                if join_msg in line:
                    found_join_msg = True
                # Logged when doing an RPC call to a node to start a game.
                elif start_game_msg in line:
                    found_start_game_msg = True

            self.assertTrue(found_join_msg,
                            "MS should've logged that {} joined".format(client.node_port))
            self.assertTrue(found_start_game_msg,
                            "MS should've logged that it notified {} of game "
                            "start".format(client.node_port))

        # Check that all clients interact with the MS.
        for client in clients:
            # Logged when doing a Join RPC call to MS.
            join_msg = "Rpc Call Context.Join to localhost:{}".format(ms_srv.port)
            found_join_msg = False
            # Logged when receiving an RPC call to start the game.
            start_game_msg = "Rpc Called Start Game to localhost:{}".format(
                ms_srv.port)
            found_start_game_msg = False
            lines = []
            with open(client.govector_log_path) as log_file:
                lines = log_file.readlines()

            for line in lines[1::2]:
                if join_msg in line:
                    found_join_msg = True
                elif start_game_msg in line:
                    found_start_game_msg = True

            self.assertTrue(found_join_msg,
                            "Node {} should log that it contacted the MS for "
                            "joining".format(client.node_port))
            self.assertTrue(found_start_game_msg,
                            "Node {} should log that the MS said the game was "
                            "starting".format(client.node_port))

            # Check that all clients interact with each other.
            other_clients = [c for c in clients if c != client]
            for other_client in other_clients:
                other_address = "localhost:{}".format(other_client.node_port)
                # Logged when sending an interval update message to peers.
                send_interval_regex = re.compile(
                    "Sending: Interval update.* at ip (.*)]")
                found_send_interval_msg = False
                # TODO: This regex needs to be tightened so it only matches
                #       interval update packets.
                # Logged when receiving an interval update msg from a peer.
                recv_interval_regex = re.compile(
                    'Received packet from .* {.*"Node".*"Ip":"(.*)","Curr')
                found_recv_interval_msg = False
                for line in lines[1::2]:
                    send_interval_msg_match = send_interval_regex.match(line)
                    recv_interval_msg_match = recv_interval_regex.match(line)
                    if (send_interval_msg_match and
                        send_interval_msg_match.group(1) == other_address):
                        found_send_interval_msg = True
                    elif (recv_interval_msg_match and
                          recv_interval_msg_match.group(1) == other_address):
                        found_recv_interval_msg = True

                self.assertTrue(found_send_interval_msg,
                                "Node {} should log that it sent an interval "
                                "update to {}".format(client.node_port,
                                                        other_client.node_port))
                self.assertTrue(found_recv_interval_msg,
                                "Node {} should log that it received an "
                                "interval update from {}".format(client.node_port,
                                                                 other_client.node_port))

if __name__ == "__main__":
    unittest.main()
