#!/usr/bin/env python2

import os
import sys
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class MultiClientFailureTest(common.TestCase):
    def test_multi_client_failure(self):
        """A leader client (1) is started followed by two normal clients (2, 3, 4).
        2 and 3 fail. 1 should stay as leader and broadcast that 2 and 3 failed to
        all other clients, and 4 should stay as a client.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        common.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 4)

        common.sleep(common.MatchMakingServer.GAME_START_TIMEOUT)

        # Kill client 2 and 3.
        common.sleep(2)
        client2 = clients[1]
        client2.kill()
        client3 = clients[2]
        client3.kill()

        # Wait for (unexpected, but potential) leader re-election to occur.
        common.sleep(8)

        leader = clients[0]
        with open(leader.local_log_path) as log_file:
            found_leader_msg = False
            found_node_msg = False
            for line in log_file:
                if "Im a leader" in line:
                    found_leader_msg = True
                    continue
                elif "Im a node" in line:
                    found_node_msg = True
                    continue
            self.assertFalse(found_node_msg,
                             "Leader should never have been a node")
            self.assertTrue(found_leader_msg,
                            "Leader should always been such")

        client4 = clients[3]
        with open(client4.local_log_path) as log_file:
            found_leader_msg = False
            found_node_msg = False
            for line in log_file:
                if "Im a leader" in line:
                    found_leader_msg = True
                    continue
                elif "Im a node" in line:
                    found_node_msg = True
                    continue
            self.assertTrue(found_node_msg,
                            "Client 4 should have been a node at some point")
            self.assertFalse(found_leader_msg,
                             "Client 4 should never have been a leader")

        # TODO: We should check that the leader informed the other clients that
        #       clients 2 and 3 failed.

if __name__ == "__main__":
    unittest.main()
